package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

var (
	CPUUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_cpu_usage_percent",
		Help: "Total CPU usage percent",
	})

	MemoryUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_memory_usage_percent",
		Help: "Used memory percent",
	})

	DiskUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "system_disk_usage_percent",
		Help: "Disk usage percent of root filesystem",
	})
)

func Init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(CPUUsage, MemoryUsage, DiskUsage)

	// Start background updater
	go func() {
		for {
			update()
			time.Sleep(5 * time.Second)
		}
	}()
}

func update() {
	// CPU
	if percents, err := cpu.Percent(0, false); err == nil && len(percents) > 0 {
		CPUUsage.Set(percents[0])
	}

	// Memory
	if vm, err := mem.VirtualMemory(); err == nil {
		MemoryUsage.Set(vm.UsedPercent)
	}

	// Disk
	if du, err := disk.Usage("/"); err == nil {
		DiskUsage.Set(du.UsedPercent)
	}
}
