package main
import (
    "fmt"
	"strings"
    "os/exec"
	"flag"
	"bytes"
	"strconv"
	"os"
	"sync"
	"io/ioutil"
	"net/http"
	pipe "github.com/b4b4r07/go-pipe"
	"github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)


var (
    masterIP = flag.String("masterIP", "192.168.0.1", "Iris Master LocalIP run this exporter");
    irisBinPath = flag.String("irisBinPath", "/home/iris/IRIS/bin/", "Iris Command Binary Path");
	//sedfile = flag.String("sedfile", "sedcommand.file", "Complicated sed command");
	mpsLabelStr = []string{"node_ip","role","abn","mid","name","desc","mode","pid","cmd","sta","uptime"}
	ntopLabelStr = []string{"node_num","sys_status","adm_status","update_time","node_ip"}
)


type Collector struct {
    sync.Mutex
    mux                       *http.ServeMux                      
    mpsStatus                 *prometheus.GaugeVec
	ntopStatus                *prometheus.GaugeVec
	ntopCpu                   *prometheus.GaugeVec
	ntopLoadAvg               *prometheus.GaugeVec
	ntopMemp                  *prometheus.GaugeVec
	ntopMemf                  *prometheus.GaugeVec
	ntopDisk                  *prometheus.GaugeVec
	targetScrapeRequestErrors prometheus.Counter
	mpsDesc                   *prometheus.Desc
	ntopStatusDesc            *prometheus.Desc
	ntopCpuDesc               *prometheus.Desc
	ntopLoadAvgDesc           *prometheus.Desc
	ntopMempDesc              *prometheus.Desc
	ntopMemfDesc              *prometheus.Desc
	ntopDiskDesc              *prometheus.Desc
	Registry                  *prometheus.Registry
	options                   Options
}

type Options struct {
	Registry      *prometheus.Registry
}

func (c *Collector) scrapeHandler(w http.ResponseWriter, r *http.Request) {
    c.mpsStatus.Reset();
	c.ntopStatus.Reset();
    c.ntopCpu.Reset();
	c.ntopLoadAvg.Reset();
	c.ntopMemp.Reset();
	c.ntopMemf.Reset();
	c.ntopDisk.Reset();
	c.GetMPSMaster();
	c.GetMPSSub();
	c.GetNodeStatus();
	promhttp.HandlerFor(
		c.options.Registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError},
	).ServeHTTP(w, r)
}


func NewIrisMetricExporter(opts Options) (*Collector, error) {

	c := &Collector{
	    options: opts,
		targetScrapeRequestErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "iris",
			Name:      "target_scrape_request_errors_total",
			Help:      "Errors in requests to the exporter",
		}),
	}
	
    c.mpsDesc = prometheus.NewDesc(
                 "iris_process_status",
                 "iris Process Status",
                 mpsLabelStr, nil,
    )
	
	c.mpsStatus = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "iris_process_status",
				Help: "iris Process Status",
	}, mpsLabelStr)
	
    c.ntopStatusDesc = prometheus.NewDesc(
                  "iris_node_status",
                  "iris Node Status",
                  ntopLabelStr, nil,
    )
	c.ntopCpuDesc = prometheus.NewDesc(
                  "iris_node_cpu_usage",
                  "iris Node CPU Usage",
                  ntopLabelStr, nil,
    )
	
	c.ntopLoadAvgDesc = prometheus.NewDesc(
                  "iris_node_loadavg_usage",
                  "iris Node LoadAvg Status",
                  ntopLabelStr, nil,
    )
	
	c.ntopMempDesc = prometheus.NewDesc(
                  "iris_node_memp_usage",
                  "iris Node MemP usage",
                  ntopLabelStr, nil,
    )
	
	c.ntopMemfDesc = prometheus.NewDesc(
                  "iris_node_memf_usage",
                  "iris Node MemF usage",
                  ntopLabelStr, nil,
    )
	
	c.ntopDiskDesc = prometheus.NewDesc(
                  "iris_node_disk_usage",
                  "iris Node Disk usage",
                  ntopLabelStr, nil,
    )


	c.ntopStatus = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "iris_node_status",
				Help: "iris Node Status",
	}, ntopLabelStr)

    c.ntopCpu = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "iris_node_cpu_usage",
				Help: "iris Node CPU Usage",
	}, ntopLabelStr)
	
	c.ntopLoadAvg = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "iris_node_loadavg_usage",
				Help: "iris Node LoadAvg Usage",
	}, ntopLabelStr)
	
	c.ntopMemp = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "iris_node_memp_usage",
				Help: "iris Node MemP Usage",
	}, ntopLabelStr)
	
	c.ntopMemf = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "iris_node_memf_usage",
				Help: "iris Node MemF Usage",
	}, ntopLabelStr)
	
	c.ntopDisk = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "iris_node_disk_usage",
				Help: "iris Node Disk Usage",
	}, ntopLabelStr)

	
	if c.options.Registry != nil {
		c.options.Registry.MustRegister(c.targetScrapeRequestErrors)
		c.options.Registry.MustRegister(c.mpsStatus)
		c.options.Registry.MustRegister(c.ntopStatus)
		c.options.Registry.MustRegister(c.ntopCpu)
		c.options.Registry.MustRegister(c.ntopLoadAvg)
		c.options.Registry.MustRegister(c.ntopMemp)
		c.options.Registry.MustRegister(c.ntopMemf)
		c.options.Registry.MustRegister(c.ntopDisk)
	}
    c.mux = http.NewServeMux()
	c.mux.HandleFunc("/metrics", c.scrapeHandler)
	c.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`ok`))
	})

	return c, nil
}


func (c *Collector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.mux.ServeHTTP(w, r)
}

// Describe outputs NCHF metric descriptions.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.mpsDesc
	ch <- c.ntopStatusDesc
	ch <- c.ntopCpuDesc
	ch <- c.ntopLoadAvgDesc
	ch <- c.ntopMempDesc
	ch <- c.ntopMemfDesc
	ch <- c.ntopDiskDesc
	ch <- c.targetScrapeRequestErrors.Desc()

}


func (c *Collector) DerefString(s *string) string {
    if s != nil {
        return *s
    }

    return ""
}

func (c *Collector) MpsMasterParser(result []byte) [][]string {
	var (
		labels [][]string
    )

	for _, line := range strings.Split(string(result), "\n") {	
		if len(line) == 0 {
			continue
		}
		//fmt.Println("MpsMasterParser Check Data : " + line)
		bufferStr := strings.Split(line, "|")
		label := []string{c.DerefString(masterIP), 
		                  strings.TrimSpace(bufferStr[0]), 
						  strings.TrimSpace(bufferStr[1]), 
						  strings.TrimSpace(bufferStr[2]), 
						  strings.TrimSpace(bufferStr[3]), 
						  strings.TrimSpace(bufferStr[4]), 
						  strings.TrimSpace(bufferStr[5]), 
						  strings.TrimSpace(bufferStr[6]), 
						  strings.TrimSpace(bufferStr[7]), 
						  strings.TrimSpace(bufferStr[8]), 
						  strings.TrimSpace(bufferStr[9])}
		labels= append(labels,label)
	}
	return labels
}

func (c *Collector) GetMPSMaster() {
    result, err := c.Execute("mps-master")
    if err != nil {
            fmt.Fprintln(os.Stderr, "There was an error in running iris command ", err)
            c.targetScrapeRequestErrors.Add(1);
    } else if strings.Contains(string(result), "EHD is not working") {
            fmt.Fprintln(os.Stderr, "There was an error in running iris command : EHD is not working ", err)
            c.targetScrapeRequestErrors.Add(1);   
    } else {
	    labels := c.MpsMasterParser(result)
	    for _, label := range labels { 
			var targetMetric float64 = 0 
			
			if(label[1] == "OK") {
			    targetMetric = 1
			}
			//fmt.Println("GetMPSMaster Check Data : " + label[1])
			c.mpsStatus.WithLabelValues(label[0], label[1], label[2], label[3], 
			                          label[4], label[5], label[6], label[7], 
									  label[8], label[9], label[10]).Set(targetMetric)
		}
    }
}

func (c *Collector) MpsSubParser(result []byte) [][]string {
	var (
		labels [][]string
		subIP string
    )
    //fmt.Println("MpsSubParser Check Data 1 : " + string(result))
	for _, line := range strings.Split(string(result), "\n") {	
		if len(line) == 0 {
			continue
		}
		
		//fmt.Println("MpsSubParser Check Data 2 : " + line)
		
		if len(line) < 16 {
		    subIP = strings.TrimSpace(line)
			continue
		}
		
		bufferStr := strings.Split(line, "|")
		label := []string{subIP, 
		                  strings.TrimSpace(bufferStr[0]), 
						  strings.TrimSpace(bufferStr[1]), 
						  strings.TrimSpace(bufferStr[2]), 
						  strings.TrimSpace(bufferStr[3]), 
						  strings.TrimSpace(bufferStr[4]), 
						  strings.TrimSpace(bufferStr[5]), 
						  strings.TrimSpace(bufferStr[6]), 
						  strings.TrimSpace(bufferStr[7]), 
						  strings.TrimSpace(bufferStr[8]), 
						  strings.TrimSpace(bufferStr[9])}
		labels= append(labels,label)
	}
	return labels
}

func (c *Collector) GetMPSSub() {
    result, err := c.Execute("mps-sub")
    if err != nil {
            fmt.Fprintln(os.Stderr, "There was an error in running iris command ", err)
            c.targetScrapeRequestErrors.Add(1);
    } else if strings.Contains(string(result), "EHD is not working") {
            fmt.Fprintln(os.Stderr, "There was an error in running iris command : EHD is not working ", err)
            c.targetScrapeRequestErrors.Add(1);            
    } else {
	    labels := c.MpsSubParser(result)
	    for _, label := range labels { 
			var targetMetric float64 = 0 
			
			if(label[1] == "OK") {
			    targetMetric = 1
			}
			
			c.mpsStatus.WithLabelValues(label[0], label[1], label[2], label[3], 
			                          label[4], label[5], label[6], label[7], 
									  label[8], label[9], label[10]).Set(targetMetric)
		}
    }

}

func (c *Collector) NodeStatusParser(result []byte) [][]string {
	var (
		labels [][]string
    )

	for _, line := range strings.Split(string(result), "\n") {	
		if len(line) == 0 {
			continue
		}
		
		//fmt.Println("NodeStatusParser Check Data : " + line)
		
		bufferStr := strings.Split(line, ",")
		label := []string{strings.TrimSpace(strings.Replace(bufferStr[0], "NODE:", "", 1)), 
						  strings.TrimSpace(bufferStr[1]), 
						  strings.TrimSpace(bufferStr[2]), 
						  strings.TrimSpace(bufferStr[3]), 
						  strings.TrimSpace(bufferStr[4]), 
						  strings.TrimSpace(bufferStr[5]), 
						  strings.TrimSpace(bufferStr[6]), 
						  strings.TrimSpace(bufferStr[7]), 
						  strings.TrimSpace(bufferStr[8]), 
						  strings.TrimSpace(bufferStr[9])}
		labels= append(labels,label)
	}
	return labels
}

func (c *Collector) GetNodeStatus() {
	result, err := c.Execute("ntop")
    if err != nil {
		fmt.Fprintln(os.Stderr, "There was an error in running iris command ", err)
		c.targetScrapeRequestErrors.Add(1);
    } else {
	    labels := c.NodeStatusParser(result)
	    for _, label := range labels { 
			var targetMetric float64 = 0 
			
			if(label[1] == "VALID") {
			    targetMetric = 1
			} else if(label[1] == "WAIT_RETRY") {
			    targetMetric = -1
			}
			
			c.ntopStatus.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(targetMetric)
			
			cpuVal, err := strconv.ParseFloat(label[5], 64)
			if err != nil {
			    c.ntopCpu.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(0)
			} else {
			    c.ntopCpu.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(cpuVal)
			}
			
			loadVal, err := strconv.ParseFloat(label[6], 64)
			if err != nil {
			    c.ntopLoadAvg.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(0)
			} else {
			    c.ntopLoadAvg.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(loadVal)
			}			

			mempVal, err := strconv.ParseFloat(label[7], 64)
			if err != nil {
			    c.ntopMemp.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(0)
			} else {
			    c.ntopMemp.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(mempVal)
			}

			memfVal, err := strconv.ParseFloat(label[8], 64)
			if err != nil {
			    c.ntopMemf.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(0)
			} else {
			    c.ntopMemf.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(memfVal)
			}

			diskVal, err := strconv.ParseFloat(label[9], 64)
			if err != nil {
			    c.ntopDisk.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(0)
			} else {
			    c.ntopDisk.WithLabelValues(label[0], label[1], label[2], label[3], 
			                           label[4]).Set(diskVal)
			}
		}
    }
}

func (c *Collector) Execute(commandType string) ([]byte, error) {
	var resValue bytes.Buffer
	//fmt.Println("Check Command: " + commandType)
	if(commandType == "mps-master") {
	    if err := pipe.Command(&resValue,
        	exec.Command(c.DerefString(irisBinPath) + "mps"),
        	exec.Command("sed", "-n", "/|/p"),
        ); err != nil {
        	return nil, err
        }
    } else if(commandType == "mps-sub") {
	    if err := pipe.Command(&resValue,
        	exec.Command(c.DerefString(irisBinPath) + "cmd", "mps"),
        	exec.Command("sed", "-e", "/</d;/--/d;/ABN   MID    NAME         DESC                      GROUP       MODE    PID    CMD    STA         TIME/d"),
        ); err != nil {
        	return nil, err
        }
    } else {
        if err := pipe.Command(&resValue,
        	exec.Command(c.DerefString(irisBinPath) + "Admin/NodeList"),
        	exec.Command("sed", "-n", "/NODE:/p"),
        ); err != nil {
        	return nil, err
        }
    }	
	
	if bytes, err := ioutil.ReadAll(&resValue); err != nil {
        return nil, err
    } else{
    	return bytes, nil
    }
	//mps | sed -n '/|/p'
	//cmd mps | sed -f sedcommand.file 
	//ntop | sed -n '/NODE:/p'

}

func main() {
	listen := flag.String("listen", ":9202", "listen on")
	flag.Parse()
	
    registry := prometheus.NewRegistry()
	cltr, err := NewIrisMetricExporter(
		Options{
			Registry:      registry,
		},
	)
	
	if err != nil {
		log.Fatal(err)
	} 
	
    log.Fatal(http.ListenAndServe(*listen, cltr))
}

