package troll

import (
	"fmt"
	"os"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const TrollAnnotationKey = "bridge-troll.monitoring.pusher.com/actual-config-hash"

var staleMetric = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "troll_files_stale",
	Help: "1 if watch files are stale, 0 otherwise",
})

type BridgeTroll struct {
	WatchList    []string
	Client       *kubernetes.Clientset
	PodName      string
	PodNamespace string
	Hash         string
}

func NewBridgeTroll(watchList []string) (troll *BridgeTroll, err error) {
	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")
	if podName == "" || podNamespace == "" {
		err = fmt.Errorf("init failed: POD_NAME or POD_NAMESPACE environment variables not set")
		return
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return
	}
	troll = &BridgeTroll{
		WatchList:    watchList,
		Client:       client,
		PodName:      podName,
		PodNamespace: podNamespace,
	}
	troll = &BridgeTroll{
		WatchList:    watchList,
		Client:       nil,
		PodName:      "",
		PodNamespace: "",
	}
	return
}

func (t *BridgeTroll) Start() (*sync.WaitGroup, error) {
	pod, err := t.Client.CoreV1().Pods(t.PodNamespace).Get(t.PodName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod details: %s", err)
	}

	if pod.Annotations == nil {
		pod.Annotations = make(map[string]string)
	}
	if h, ok := pod.Annotations[TrollAnnotationKey]; ok {
		t.Hash = h
	} else {
		hash, err := hashFiles(t.WatchList)
		if err != nil {
			return nil, fmt.Errorf("failed to hash watchfile list: %s", err)
		}
		pod.Annotations[TrollAnnotationKey] = hash
		pod, err = t.Client.CoreV1().Pods(t.PodNamespace).Update(pod)
		if err != nil {
			return nil, fmt.Errorf("unable to annotate pod: %s", err)
		}
		t.Hash = hash
	}
	http.Handle("/metrics", promhttp.Handler())
	// TODO: Metrics port configurable
	go http.ListenAndServe(":2112", nil)

	hash, err := hashFiles(t.WatchList)
	if err != nil {
		return nil, fmt.Errorf("unable to hash files: %s", err)
	}
	t.Hash = hash
	sync := &sync.WaitGroup{}
	sync.Add(1)
	go t.watch(sync)
	return sync, nil
}

func (t *BridgeTroll) watch(sync *sync.WaitGroup) {
	defer sync.Done()
	var status bool
	var err error
	for {
		status, err = t.check()
		if err != nil {
			fmt.Printf("Error in goroutine: %s", err)
			panic(err)
		}
		if !status {
			staleMetric.Set(1)
		} else {
			staleMetric.Set(0)
		}
		time.Sleep(2 * time.Second)
	}
}

func (t *BridgeTroll) check() (bool, error) {
	currentHash, err := hashFiles(t.WatchList)
	if err != nil {
		return false, fmt.Errorf("unable to verify file hash: %s", err)
	}
	return (currentHash == t.Hash), nil
}
