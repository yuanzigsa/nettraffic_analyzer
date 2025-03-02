package main

import (
    "sync"
    "time"
    "github.com/google/gopacket"
    "github.com/google/gopacket/layers"
)

func packetProcessor(handle gopacket.PacketDataSource, wg *sync.WaitGroup) {
    defer wg.Done()

    source := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
    for packet := range source.Packets() {
        // 解析IP层
        ipLayer := packet.Layer(layers.LayerTypeIPv4)
        if ipLayer == nil {
            continue
        }
        ip := ipLayer.(*layers.IPv4)

        // 解析传输层
        var srcPort, dstPort uint16
        var protocol uint8
        
        tcpLayer := packet.Layer(layers.LayerTypeTCP)
        if tcpLayer != nil {
            tcp := tcpLayer.(*layers.TCP)
            srcPort = uint16(tcp.SrcPort)
            dstPort = uint16(tcp.DstPort)
            protocol = 6
        } else {
            udpLayer := packet.Layer(layers.LayerTypeUDP)
            if udpLayer != nil {
                udp := udpLayer.(*layers.UDP)
                srcPort = uint16(udp.SrcPort)
                dstPort = uint16(udp.DstPort)
                protocol = 17
            } else {
                continue
            }
        }

        // 创建流标识
        flowKey := FlowKey{
            SrcIP:    ip.SrcIP.String(),
            DstIP:    ip.DstIP.String(),
            SrcPort:  srcPort,
            DstPort:  dstPort,
            Protocol: protocol,
        }

        // 更新统计信息
        flowMapLock.Lock()
        stats, exists := flowMap[flowKey]
        if !exists {
            stats = &FlowStats{}
            flowMap[flowKey] = stats
        }
        stats.Bytes += uint64(len(packet.Data()))
        stats.Packets++
        stats.LastSeen = time.Now()
        flowMapLock.Unlock()
    }
} 