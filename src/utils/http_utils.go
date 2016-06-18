package utils

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func ReadAndUnmarshalObject(r io.ReadCloser, objptr interface{}) error {
	if result, err := ioutil.ReadAll(r); err != nil {
		return err
	} else if err = json.Unmarshal(result, objptr); err != nil {
		return err
	}
	return nil
}

func WriteObjectResponse(w http.ResponseWriter, obj interface{}) error {
	if result, err := json.Marshal(obj); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	} else {
		_, err := w.Write(result)
		return err
	}
}
