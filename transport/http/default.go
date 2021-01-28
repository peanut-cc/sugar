package http

import (
	"context"
	"fmt"
	"github.com/peanut-cc/sugar/errors"
	"io/ioutil"
	"net/http"
	"runtime"
)

// DefaultRequestDecoder default request decoder.
func DefaultRequestDecoder(in interface{}, req *http.Request) error {
	codec, err := RequestCodec(req)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return errors.DataLoss("DataLoss", err.Error())
	}
	defer req.Body.Close()
	if err = codec.Unmarshal(data, in); err != nil {
		return errors.InvalidArgument("CodecUnmarshal", err.Error())
	}
	return nil
}


// DefaultResponseEncoder is default response encoder.
func DefaultResponseEncoder(out interface{}, res http.ResponseWriter, req *http.Request) error {
	contentType, codec, err := ResponseCodec(req)
	if err != nil {
		return err
	}
	data, err := codec.Marshal(out)
	if err != nil {
		return err
	}
	res.Header().Set("content-type", contentType)
	res.Write(data)
	return nil
}

// DefaultErrorEncoder is default errors encoder.
func DefaultErrorEncoder(err error, res http.ResponseWriter, req *http.Request) {
	code, se := StatusError(err)
	contentType, codec, err := ResponseCodec(req)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := codec.Marshal(se)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Set("content-type", contentType)
	res.WriteHeader(code)
	res.Write(data)
}

// DefaultRecoveryHandler is default recovery handler.
func DefaultRecoveryHandler(ctx context.Context, req, err interface{}) error {
	buf := make([]byte, 65536)
	n := runtime.Stack(buf, false)
	buf = buf[:n]
	fmt.Printf("panic: %v %v\nstack: %s\n", req, err, buf)
	return errors.Unknown("Unknown", "panic triggered: %v", err)
}

