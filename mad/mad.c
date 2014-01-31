#include <mad.h>
#include <stdio.h>
#include <unistd.h>
#include <string.h>

static enum mad_flow input(void *data, struct mad_stream *stream) {
	static char buf[32*1024];
	int start;

	if (stream->buffer) {
		start = sizeof(buf) - ((char *)stream->next_frame - buf);
		memmove(buf, stream->next_frame, start);
	} else
		start = 0;

	int i = fread(buf + start, 1, sizeof(buf) - start, stdin);
	if (i <= 0) 
		return MAD_FLOW_STOP;

	//fprintf(stderr, "start %d len %d\n", start, i);

  mad_stream_buffer(stream, buf, start + i);

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
  unsigned int nchannels, nsamples;
  mad_fixed_t const *left_ch, *right_ch;

  /* pcm->samplerate contains the sampling frequency */

  nchannels = pcm->channels;
  nsamples  = pcm->length;
  left_ch   = pcm->samples[0];
  right_ch  = pcm->samples[1];

  while (nsamples--) {
    signed int sample;

    /* output sample(s) in 16-bit signed little-endian PCM */

    sample = scale(*left_ch++);
    putchar((sample >> 0) & 0xff);
    putchar((sample >> 8) & 0xff);

    if (nchannels == 2) {
      sample = scale(*right_ch++);
      putchar((sample >> 0) & 0xff);
      putchar((sample >> 8) & 0xff);
    }
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


int main() {
	struct mad_decoder decoder;

	mad_decoder_init(&decoder, 0, input, 0, 0, output, error, 0);
	mad_decoder_run(&decoder, MAD_DECODER_MODE_SYNC);
	mad_decoder_finish(&decoder);
}

