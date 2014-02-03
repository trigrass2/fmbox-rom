
package mad

// #cgo CFLAGS: -I.
// #cgo 386 LDFLAGS: libmad-386.a
// #cgo arm LDFLAGS: libmad-arm.a
// extern int run(void *data);
import "C"
import "unsafe"
import "reflect"
import "io"

type Decoder struct {
	R io.Reader
	W io.Writer
}

func convBuf(_buf unsafe.Pointer, size C.int) (buf []byte) {
	bufhdr := (*reflect.SliceHeader)((unsafe.Pointer(&buf)))
	bufhdr.Cap = int(size)
	bufhdr.Len = int(size)
	bufhdr.Data = uintptr(_buf)
	return
}

//export InputCb
func InputCb(_dec, _buf unsafe.Pointer, size C.int, ret *C.int) {
	dec := (*Decoder)(_dec)
	buf := convBuf(_buf, size)
	n, err := dec.R.Read(buf)
	if err != nil {
		n = -1
	}
	*ret = C.int(n)
}

//export OutputCb
func OutputCb(_dec, _buf unsafe.Pointer, size C.int, ret *C.int) {
	dec := (*Decoder)(_dec)
	buf := convBuf(_buf, size)
	n, err := dec.W.Write(buf)
	if err != nil {
		n = -1
	}
	*ret = C.int(n)
}

func (d *Decoder) Run() error {
	r := C.run(unsafe.Pointer(d))
	if r < 0 {
		return io.ErrClosedPipe
	}
	return nil
}

func NewDecoder() *Decoder {
	return &Decoder{}
}

