package webhook

import (
	mlog "github.com/maxwell92/log"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	glog = mlog.Log
)

// MyConf crd 自定义的config
type MyConf struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Data              map[string]string `json:"data,omitempty" protobuf:"bytes,2,rep,name=data"`
}

// operators
var (
	OpAdd     = "add"
	OpRemove  = "remove"
	OpReplace = "replace"
)

// Op Path
var (
	PathAnnotations = "/metadata/annotations" //map, 单个key的话,值是string
	PathLabels      = "/metadata/labels"      //map, 单key, 值是string
	PathContainers  = "/spec/containers"      // slice
)

// 是否需要注入标识
var (
	INJECTEDKEY         = "yulibaozi/injected"   // INJECTEDKEY 标识已经被注入过
	NOINJECTEDKEY       = "yulibaozi/noinjected" // NOINJECTEDKEY 标识为不需要注入
	injectedAnnotations = map[string]string{INJECTEDKEY: "true", NOINJECTEDKEY: "true"}
)

// patchOperation patch 操作定义
type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

var (
	name, namespace string = "sidecarconf", "default"
)
var (
	defaultContainer = corev1.Container{Name: "webhook-added-container", Image: "yulibaozi/web:1.0"}
)

type admitFunc func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)
