
#include <mad.h>
#include <stdio.h>
#include <unistd.h>
#include <string.h>
#include <stdint.h>
#include "_cgo_export.h"

void InputCb(void *, void *, int, int*);
void OutputCb(void *, void *, int, int*);

static enum mad_flow input(void *data, struct mad_stream *stream) {
	static char buf[32*1024];
	int start;

	if (stream->buffer) {
		start = sizeof(buf) - ((char *)stream->next_frame - buf);
		memmove(buf, stream->next_frame, start);
	} else
		start = 0;

	int r;
	InputCb(data, buf + start, sizeof(buf) - start, &r);
	if (r < 0) 
		return MAD_FLOW_STOP;

	//fprintf(stderr, "start %d len %d\n", start, i);

  mad_stream_buffer(stream, buf, start + r);

  return MAD_FLOW_CONTINUE;
}

static inline int16_t scale(mad_fixed_t sample) {
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

	int16_t out[sizeof(pcm->samples[0])/sizeof(pcm->samples[0][0])*2];
	int outp = 0;

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

	int r;
	OutputCb(data, out, outp, &r);

	// error
	if (r < 0)
		return MAD_FLOW_BREAK;

  return MAD_FLOW_CONTINUE;
}

static enum mad_flow error(void *data, struct mad_stream *stream, struct mad_frame *frame) {
  return MAD_FLOW_CONTINUE;
}

int run(void *data) {
	struct mad_decoder decoder; 
	int r;
	mad_decoder_init(&decoder, data, input, 0, 0, output, error, 0);
	r = mad_decoder_run(&decoder, MAD_DECODER_MODE_SYNC);
	mad_decoder_finish(&decoder);
	return r;
}

