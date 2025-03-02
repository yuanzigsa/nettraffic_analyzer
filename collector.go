package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "time"
)

type FlowRecord struct {
    Timestamp   int64   `json:"timestamp"`
    PortName    string  `json:"port_name"`
    SrcIP       string  `json:"src_ip"`
    SrcPort     uint16  `json:"src_port"`
    DstIP       string  `json:"dst_ip"`
    DstPort     uint16  `json:"dst_port"`
    Protocol    uint8   `json:"protocol"`
    Bandwidth   float64 `json:"bandwidth"`
}

func statsCollector() {
    // 创建输出目录
    outputDir := "flow_stats"
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        log.Fatalf("Failed to create output directory: %v", err)
    }

    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        flowMapLock.Lock()
        records := make([]FlowRecord, 0, len(flowMap))
        
        now := time.Now()
        for key, stats := range flowMap {
            // 计算带宽 (bits per second)
            bandwidth := float64(stats.Bytes*8) / 300 // 300秒 = 5分钟

            record := FlowRecord{
                Timestamp:   now.Unix(),
                PortName:    "enp4s0f0",
                SrcIP:      key.SrcIP,
                SrcPort:    key.SrcPort,
                DstIP:      key.DstIP,
                DstPort:    key.DstPort,
                Protocol:   key.Protocol,
                Bandwidth:  bandwidth,
            }
            records = append(records, record)
        }

        // 清空统计数据
        flowMap = make(map[FlowKey]*FlowStats)
        flowMapLock.Unlock()

        // 生成文件名
        filename := filepath.Join(outputDir, fmt.Sprintf("flow_stats_%s.json", 
            now.Format("2006-01-02_15-04-05")))

        // 将数据写入文件
        file, err := os.Create(filename)
        if err != nil {
            log.Printf("Failed to create output file: %v", err)
            continue
        }

        encoder := json.NewEncoder(file)
        encoder.SetIndent("", "    ")
        if err := encoder.Encode(records); err != nil {
            log.Printf("Failed to write data: %v", err)
        }
        
        file.Close()
        log.Printf("Wrote %d records to %s", len(records), filename)
    }
} 