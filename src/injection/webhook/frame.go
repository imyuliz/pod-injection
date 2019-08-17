package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/api/admission/v1beta1"
)

func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		errstr := "empty body"
		glog.Errorln(errstr)
		http.Error(w, errstr, http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		err := fmt.Errorf("Content-Type=%s, expect application/json, invalid Content-Type, expect `application/json`", contentType)
		glog.Errorf(err.Error())
		http.Error(w, err.Error(), http.StatusUnsupportedMediaType)
		return
	}

	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Errorf("can't decode body: %v", err)
		reviewResponse = toAdmissionResponse(err)
	} else {
		glog.Infof("input admit operator, URL: %s", r.URL.Path)
		reviewResponse = admit(ar)
		glog.Infoln("finished Admit")
	}

	admissionReview := v1beta1.AdmissionReview{}
	if reviewResponse != nil {
		admissionReview.Response = reviewResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		err := fmt.Errorf("Can't encode response: %v", err)
		glog.Errorf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	glog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		err = fmt.Errorf("Can't write response: %v", err)
		glog.Errorf(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
