package main

import (
    "log"
    "sync"
    "time"
    "github.com/google/gopacket"
    "github.com/google/gopacket/afpacket"
    "golang.org/x/sys/unix"
)

// 会话信息结构体
type FlowKey struct {
    SrcIP    string
    DstIP    string
    SrcPort  uint16
    DstPort  uint16
    Protocol uint8
}

type FlowStats struct {
    Bytes    uint64
    Packets  uint64
    LastSeen time.Time
}

// 全局变量
var (
    flowMap     = make(map[FlowKey]*FlowStats)
    flowMapLock sync.RWMutex
)

// 添加pageSize定义
const pageSize = unix.Getpagesize()

func main() {
    // 创建抓包句柄
    handle, err := afpacket.NewTPacket(
        afpacket.OptInterface("enp4s0f0"),
        afpacket.OptFrameSize(65536),
        afpacket.OptBlockSize(pageSize * 128),  // 使用pageSize
        afpacket.OptNumBlocks(128),
        afpacket.OptPollTimeout(1000),  // 添加超时设置
    )
    if err != nil {
        log.Fatal(err)
    }
    defer handle.Close()

    // 启动多个工作协程
    var wg sync.WaitGroup
    for i := 0; i < 4; i++ {
        wg.Add(1)
        go packetProcessor(handle, &wg)
    }

    // 启动统计协程
    go statsCollector()

    wg.Wait()
} 