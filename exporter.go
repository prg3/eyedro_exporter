package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"net/http"
	"fmt"
	"encoding/json"
	"strconv"
	"os"
	"time"
	"log"
	"io/ioutil"
)

type metrics struct {
	power_factor  prometheus.GaugeVec
	voltage  prometheus.GaugeVec
	current  prometheus.GaugeVec
	power  prometheus.GaugeVec
}

type energy_leg struct {
	power_factor int
	voltage int
	current int
	power int
}

type EnergyData struct {
	Data [][]int `json:"data"`
}


func getJSON( ip string ) []byte {
	url := fmt.Sprintf("http://%s:8080/getdata", ip)

	spaceClient := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "eyedro-exporter")

	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	return (body)

}
func getMetrics(ip string) [20]energy_leg {


	// jsonData := `{"data":[[812,12638,15400,1580],[820,12634,13880,1437]]}`
	// data := energy{}
	var tmp EnergyData

	var data [20]energy_leg

    if err := json.Unmarshal(getJSON(ip), &tmp); err != nil {
		fmt.Println(err)
    }


    // if err := json.Unmarshal([]byte(jsonData), &tmp); err != nil {
	// 	fmt.Println(err)
    // }


	for legArray := range (tmp.Data) {
		var leg energy_leg
		leg.power_factor = tmp.Data[legArray][0]
		leg.voltage = tmp.Data[legArray][1]
		leg.current = tmp.Data[legArray][2]
		leg.power = tmp.Data[legArray][3]
		data[legArray] = leg
	}

	fmt.Println(data)
	return(data)
}

func updateMetrics ( eyedro_ip ) {

	power_factor := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "eyedro_power_factor",
		Help:        "Power Factor.",
	},
	[]string{
		"leg",
	})

	voltage := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "eyedro_voltage",
		Help:        "Voltage, in milivolts.",
	},
	[]string{
		"leg",
	})

	current := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "eyedro_current",
		Help:        "Current in milli-amps.",
	},
	[]string{
		"leg",
	})

	power := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:        "eyedro_power",
		Help:        "Power, measured in Watts.",
	},
	[]string{
		"leg",
	})

	prometheus.MustRegister(power_factor)
	prometheus.MustRegister(voltage)
	prometheus.MustRegister(current)
	prometheus.MustRegister(power)



	energyData := getMetrics( eyedro_ip )
	for leg := range energyData {
		if energyData[leg].voltage != 0 {
			leg_str := strconv.Itoa(leg)
			power_factor.WithLabelValues(leg_str).Set(float64(energyData[leg].power_factor))
			voltage.WithLabelValues(leg_str).Set(float64(energyData[leg].voltage))
			current.WithLabelValues(leg_str).Set(float64(energyData[leg].current))
			power.WithLabelValues(leg_str).Set(float64(energyData[leg].power))
		}
	}
}

func main () {
	val, ok := os.LookupEnv("PORT")
	var listen_port string
	if ok {
		listen_port = val
	} else {
		listen_port = "8080"
	}

	ip, ok := os.LookupEnv("EYEDRO_IP")

	var eyedro_ip string

	if ok {
		eyedro_ip = ip
	} else {
		panic("Must set IP of Eyedro")
		eyedro_ip = "192.168.0.1"
	}

	fmt.Println(eyedro_ip)

	// eyedro_ip := os.Getenv("EYEDRO_IP")


	http.Handle("/metrics", promhttp.Handler())
	// http.ListenAndServe(":8080", nil)
	http.ListenAndServe(fmt.Sprintf(":%s", listen_port), nil)
}