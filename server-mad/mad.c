
/*
 * libmad mp3 decode server
 */

#include <mad.h>
#include <stdio.h>
#include <signal.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/wait.h>
#include <sys/types.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>

typedef struct {
	int r, w;
	char buf[1024*16];
	int first;
} decode_t ;

static enum mad_flow input(void *data, struct mad_stream *stream) {
	decode_t *d = (decode_t *)data;
	int start;

	if (stream->buffer) {
		start = sizeof(d->buf) - ((char *)stream->next_frame - d->buf);
		memmove(d->buf, stream->next_frame, start);
	} else
		start = 0;

	int i = read(d->r, d->buf + start, sizeof(d->buf) - start);
	fprintf(stderr, "input read: %d\n", i);
	if (i <= 0) {
		fprintf(stderr, "input read: %d\n", i);
		return MAD_FLOW_STOP;
	}

	//fprintf(stderr, "start %d len %d\n", start, i);

  mad_stream_buffer(stream, d->buf, start + i);

  return MAD_FLOW_CONTINUE;
}

static inline signed int scale(mad_fixed_t sample) {
  /* round */
  sample += (1L << (MAD_F_FRACBITS - 16));

  /* clip */
  if (sample >= MAD_F_ONE)
    sample = MAD_F_ONE - 1;
  else if (sample < -MAD_F_ONE)
    sample = -MAD_F_ONE;

  /* quantize */
  return sample >> (MAD_F_FRACBITS + 1 - 16);
}

enum mad_flow output(void *data, struct mad_header const *header, struct mad_pcm *pcm) {
	decode_t *d = (decode_t *)data;
  unsigned int nchannels, nsamples;
  mad_fixed_t const *left_ch, *right_ch;

	int16_t out[sizeof(pcm->samples[0])/sizeof(pcm->samples[0][0])*2];
	int outp = 0;

	if (d->first) {
		write(d->w, &header->samplerate, 4);
		fprintf(stderr, "output rate: %d\n", pcm->samplerate);
		d->first = 0;
	}

  /* pcm->samplerate contains the sampling frequency */

  nchannels = pcm->channels;
  nsamples  = pcm->length;
  left_ch   = pcm->samples[0];
  right_ch  = pcm->samples[1];

  while (nsamples--) {
    /* output sample(s) in 16-bit signed little-endian PCM */
    out[outp++] = scale(*left_ch++);

    if (nchannels == 2) {
      out[outp++] = scale(*right_ch++);
    } else {
			outp++;
		}
	}

	int r = write(d->w, out, outp*2);
	fprintf(stderr, "output write: %d\n", r);
	if (r < 0) {
		fprintf(stderr, "output failed\n");
		return MAD_FLOW_BREAK;
	}

  return MAD_FLOW_CONTINUE;
}

enum mad_flow error(void *data, struct mad_stream *stream, struct mad_frame *frame) {

	/*
  fprintf(stderr, "decoding error 0x%04x (%s)\n",
	  stream->error, mad_stream_errorstr(stream)
		);
		*/

  /* return MAD_FLOW_BREAK here to stop decoding (and propagate an error) */

  return MAD_FLOW_CONTINUE;
}

void decode_thread(int r, int w) {
	struct mad_decoder decoder;
	decode_t d;

	d.r = r;
	d.w = w;
	d.first = 1;

	fprintf(stderr, "decode start #%d\n", d.r);
	mad_decoder_init(&decoder, &d, input, 0, 0, output, error, 0);
	mad_decoder_run(&decoder, MAD_DECODER_MODE_SYNC);
	mad_decoder_finish(&decoder);
	fprintf(stderr, "decode end #%d\n", d.r);

	shutdown(d.r, SHUT_RDWR);
	close(d.r);
}

void sid_child(int signo) {  
	pid_t pid;  
	int stat;  
	while ((pid = waitpid(-1, &stat, WNOHANG)) > 0);  
}  

void do_socket() {
	int fd, val;
	struct sockaddr_in sa;
	int r;
	int port = 91;

	signal(SIGCHLD, sid_child);

	fd = socket(AF_INET, SOCK_STREAM, 0);
	val = 1;
	setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &val, sizeof(val));

	sa.sin_family = AF_INET;
	sa.sin_port = htons(port);
	sa.sin_addr.s_addr = htonl(INADDR_ANY);
	memset(&sa.sin_zero, 0, 8);

	r = bind(fd, (struct sockaddr *)&sa, sizeof(struct sockaddr));
	if (r < 0) {
		fprintf(stderr, "bind :%d failed\n", port);
		exit(1);
	}

	r = listen(fd, 5);
	if (r < 0) {
		fprintf(stderr, "listen :%d failed\n", port);
		exit(1);
	}

	for (;;) {
		struct sockaddr_in clisa;
		int sasize = sizeof(struct sockaddr);
		int fd1 = accept(fd, (struct sockaddr *)&sa, (socklen_t *)&sasize);
		if (fd1 < 0) {
			fprintf(stderr, "accept failed\n");
			exit(1);
		}

		r = fork();
		if (r == 0) {
			decode_thread(fd1, fd1);
			exit(0);
		} else if (r < 0) {
			fprintf(stderr, "fork failed\n");
			exit(1);
		}
	}
}

int main(int argc, char *argv[]) {

	if (argc > 1 && !strcmp(argv[1], "socket")) {
		do_socket();
		return 0;
	}

	decode_thread(0, 1);

	return 0;
}

