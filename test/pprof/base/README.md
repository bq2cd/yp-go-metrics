# Baseline

Collection: `go run ./cmd/pproftest/ -d test/pprof/base -t 300 -m 1`

## Agent

### CPU

Top: `go tool pprof -top agent-cpu.pprof`

```
File: agent-3011810301
Type: cpu
Time: 2026-04-16 11:14:41 MSK
Duration: 299.09s, Total samples = 3.60s ( 1.20%)
Showing nodes accounting for 3.38s, 93.89% of 3.60s total
Dropped 147 nodes (cum <= 0.02s)
      flat  flat%   sum%        cum   cum%
     0.54s 15.00% 15.00%      0.54s 15.00%  runtime.pthread_cond_signal
     0.41s 11.39% 26.39%      0.41s 11.39%  runtime.pthread_cond_wait
     0.36s 10.00% 36.39%      0.36s 10.00%  runtime.kevent
     0.35s  9.72% 46.11%      0.35s  9.72%  runtime.usleep
     0.32s  8.89% 55.00%      0.32s  8.89%  runtime.madvise
     0.21s  5.83% 60.83%      0.21s  5.83%  runtime.pthread_kill
     0.16s  4.44% 65.28%      0.16s  4.44%  syscall.syscall
     0.15s  4.17% 69.44%      0.39s 10.83%  runtime.pcvalue
     0.12s  3.33% 72.78%      0.12s  3.33%  runtime.pthread_cond_timedwait_relative_np
     0.12s  3.33% 76.11%      0.17s  4.72%  runtime.step
     0.07s  1.94% 78.06%      0.07s  1.94%  runtime.cgocall
     0.07s  1.94% 80.00%      0.08s  2.22%  runtime.typePointers.next
     0.05s  1.39% 81.39%      0.05s  1.39%  runtime.findfunc
     0.05s  1.39% 82.78%      0.05s  1.39%  runtime.readvarint (inline)
     0.04s  1.11% 83.89%      0.14s  3.89%  runtime.scanobject
     0.04s  1.11% 85.00%      0.04s  1.11%  syscall.syscall6
     0.03s  0.83% 85.83%      0.03s  0.83%  runtime.acquirem (inline)
     0.03s  0.83% 86.67%      0.03s  0.83%  runtime.greyobject
     0.03s  0.83% 87.50%      0.03s  0.83%  runtime.memclrNoHeapPointers
     0.03s  0.83% 88.33%      0.49s 13.61%  runtime.tracebackPCs
     0.02s  0.56% 88.89%      0.02s  0.56%  runtime.cheaprand (inline)
     0.02s  0.56% 89.44%      0.11s  3.06%  runtime.markroot
     0.02s  0.56% 90.00%      0.03s  0.83%  runtime.stkbucket
     0.02s  0.56% 90.56%      0.10s  2.78%  runtime.unlock2
     0.01s  0.28% 90.83%      0.11s  3.06%  runtime.(*unwinder).resolveInternal
     0.01s  0.28% 91.11%      0.51s 14.17%  runtime.callers
     0.01s  0.28% 91.39%      0.42s 11.67%  runtime.lock2
     0.01s  0.28% 91.67%      0.66s 18.33%  runtime.mProf_Malloc
     0.01s  0.28% 91.94%      0.35s  9.72%  runtime.newobject
     0.01s  0.28% 92.22%      0.19s  5.28%  runtime.newstack
     0.01s  0.28% 92.50%      0.30s  8.33%  runtime.pcdatavalue1
     0.01s  0.28% 92.78%      0.02s  0.56%  runtime.scanblock
     0.01s  0.28% 93.06%      0.23s  6.39%  runtime.setprofilebucket
     0.01s  0.28% 93.33%      0.03s  0.83%  runtime.shrinkstack
     0.01s  0.28% 93.61%      2.02s 56.11%  runtime.systemstack
     0.01s  0.28% 93.89%      0.02s  0.56%  strings.(*Builder).grow
         0     0% 93.89%      0.15s  4.17%  bufio.(*Writer).Flush
         0     0% 93.89%      0.09s  2.50%  bytes.(*Reader).WriteTo
         0     0% 93.89%      0.03s  0.83%  compress/flate.(*Writer).Close (inline)
         0     0% 93.89%      0.03s  0.83%  compress/flate.(*compressor).close
         0     0% 93.89%      0.03s  0.83%  compress/flate.(*compressor).encSpeed
         0     0% 93.89%      0.07s  1.94%  compress/flate.(*compressor).init
         0     0% 93.89%      0.02s  0.56%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0% 93.89%      0.02s  0.56%  compress/flate.(*huffmanBitWriter).writeBlockDynamic
         0     0% 93.89%      0.02s  0.56%  compress/flate.(*huffmanEncoder).generate
         0     0% 93.89%      0.09s  2.50%  compress/flate.NewWriter (inline)
         0     0% 93.89%      0.03s  0.83%  compress/flate.newDeflateFast (inline)
         0     0% 93.89%      0.04s  1.11%  compress/flate.newHuffmanBitWriter (inline)
         0     0% 93.89%      0.03s  0.83%  compress/gzip.(*Writer).Close
         0     0% 93.89%      0.09s  2.50%  compress/gzip.(*Writer).Write
         0     0% 93.89%      0.03s  0.83%  context.WithCancel
         0     0% 93.89%      0.02s  0.56%  context.WithDeadline (inline)
         0     0% 93.89%      0.02s  0.56%  context.WithDeadlineCause
         0     0% 93.89%      0.02s  0.56%  context.WithTimeout
         0     0% 93.89%      0.03s  0.83%  context.WithoutCancel (inline)
         0     0% 93.89%      0.02s  0.56%  context.withCancel (inline)
         0     0% 93.89%      0.02s  0.56%  crypto/hmac.New
         0     0% 93.89%      0.02s  0.56%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
         0     0% 93.89%      0.03s  0.83%  fmt.(*buffer).writeString (inline)
         0     0% 93.89%      0.02s  0.56%  fmt.(*fmt).fmtS
         0     0% 93.89%      0.02s  0.56%  fmt.(*fmt).padString
         0     0% 93.89%      0.03s  0.83%  fmt.(*pp).doPrintf
         0     0% 93.89%      0.02s  0.56%  fmt.(*pp).fmtString
         0     0% 93.89%      0.02s  0.56%  fmt.(*pp).printArg
         0     0% 93.89%      0.03s  0.83%  fmt.Fprintf
         0     0% 93.89%      0.02s  0.56%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*agent).Run.(*agent).launchCollector.func1
         0     0% 93.89%      0.03s  0.83%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*agent).Run.(*agent).launchReporter.func2
         0     0% 93.89%      0.03s  0.83%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*agent).doReport
         0     0% 93.89%      0.02s  0.56%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*collector).Collect
         0     0% 93.89%      0.18s  5.00%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*collector).Collect.func1
         0     0% 93.89%      0.18s  5.00%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*collector).collectFromSource
         0     0% 93.89%      0.03s  0.83%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).getSendableMetrics
         0     0% 93.89%      0.61s 16.94%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).processBatches.func1
         0     0% 93.89%      0.61s 16.94%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).reportBatch
         0     0% 93.89%      0.61s 16.94%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).reportWorker (inline)
         0     0% 93.89%      0.57s 15.83%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).SendBatch
         0     0% 93.89%      0.49s 13.61%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendSingleRequest
         0     0% 93.89%      0.52s 14.44%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendWithRetries
         0     0% 93.89%      0.49s 13.61%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendWithRetries.func1
         0     0% 93.89%      0.13s  3.61%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).setBody
         0     0% 93.89%      0.05s  1.39%  github.com/bq2cd/yp-go-metrics/internal/app/agent.runPeriodicTask
         0     0% 93.89%      0.02s  0.56%  github.com/bq2cd/yp-go-metrics/internal/app/agent/source/memstats.(*source).ReadMetrics
         0     0% 93.89%      0.16s  4.44%  github.com/bq2cd/yp-go-metrics/internal/app/agent/source/psutil.(*source).ReadMetrics
         0     0% 93.89%      0.07s  1.94%  github.com/bq2cd/yp-go-metrics/internal/app/agent/source/psutil.(*source).readCPUMetrics
         0     0% 93.89%      0.09s  2.50%  github.com/bq2cd/yp-go-metrics/internal/app/agent/source/psutil.(*source).readMemoryMetrics
         0     0% 93.89%      0.02s  0.56%  github.com/bq2cd/yp-go-metrics/internal/model.MetricSet.Upsert (inline)
         0     0% 93.89%      0.03s  0.83%  github.com/bq2cd/yp-go-metrics/internal/model.NewMetricSet (inline)
         0     0% 93.89%      0.03s  0.83%  github.com/bq2cd/yp-go-metrics/internal/model.NewMetricSetWithStrategy
         0     0% 93.89%      0.02s  0.56%  github.com/bq2cd/yp-go-metrics/pkg/hmacsigner.(*hmacSigner).Sign
         0     0% 93.89%      0.02s  0.56%  github.com/bq2cd/yp-go-metrics/pkg/log.Str (inline)
         0     0% 93.89%      0.05s  1.39%  github.com/bq2cd/yp-go-metrics/pkg/periodictask.(*timerTask).Run
         0     0% 93.89%      0.49s 13.61%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.*uint8]).Do
         0     0% 93.89%      0.03s  0.83%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.NewRetrier[go.shape.*uint8]
         0     0% 93.89%      0.02s  0.56%  github.com/ebitengine/purego.Dlsym
         0     0% 93.89%      0.08s  2.22%  github.com/ebitengine/purego.RegisterFunc.func4
         0     0% 93.89%      0.02s  0.56%  github.com/ebitengine/purego.RegisterLibFunc
         0     0% 93.89%      0.02s  0.56%  github.com/ebitengine/purego.loadSymbol (inline)
         0     0% 93.89%      0.31s  8.61%  github.com/go-resty/resty/v2.(*Client).execute
         0     0% 93.89%      0.07s  1.94%  github.com/go-resty/resty/v2.(*Client).executeBefore
         0     0% 93.89%      0.31s  8.61%  github.com/go-resty/resty/v2.(*Request).Execute
         0     0% 93.89%      0.31s  8.61%  github.com/go-resty/resty/v2.(*Request).Send (inline)
         0     0% 93.89%      0.03s  0.83%  github.com/go-resty/resty/v2.createHTTPRequest
         0     0% 93.89%      0.04s  1.11%  github.com/go-resty/resty/v2.parseRequestURL
         0     0% 93.89%      0.02s  0.56%  github.com/goccy/go-json.(*Encoder).Encode (inline)
         0     0% 93.89%      0.02s  0.56%  github.com/goccy/go-json.(*Encoder).EncodeWithOption
         0     0% 93.89%      0.02s  0.56%  github.com/goccy/go-json.(*Encoder).encodeWithOption
         0     0% 93.89%      0.02s  0.56%  github.com/goccy/go-json.Unmarshal (inline)
         0     0% 93.89%      0.02s  0.56%  github.com/goccy/go-json.unmarshal
         0     0% 93.89%      0.02s  0.56%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).Decode
         0     0% 93.89%      0.07s  1.94%  github.com/shirou/gopsutil/v4/cpu.Percent (inline)
         0     0% 93.89%      0.07s  1.94%  github.com/shirou/gopsutil/v4/cpu.PercentWithContext
         0     0% 93.89%      0.07s  1.94%  github.com/shirou/gopsutil/v4/cpu.TimesWithContext
         0     0% 93.89%      0.07s  1.94%  github.com/shirou/gopsutil/v4/cpu.allCPUTimes
         0     0% 93.89%      0.07s  1.94%  github.com/shirou/gopsutil/v4/cpu.percentUsedFromLastCallWithContext
         0     0% 93.89%      0.02s  0.56%  github.com/shirou/gopsutil/v4/internal/common.NewLibrary
         0     0% 93.89%      0.09s  2.50%  github.com/shirou/gopsutil/v4/mem.VirtualMemory (inline)
         0     0% 93.89%      0.09s  2.50%  github.com/shirou/gopsutil/v4/mem.VirtualMemoryWithContext
         0     0% 93.89%      0.04s  1.11%  github.com/shirou/gopsutil/v4/mem.getHwMemsize (inline)
         0     0% 93.89%      0.02s  0.56%  golang.org/x/sync/errgroup.(*Group).Go
         0     0% 93.89%      0.84s 23.33%  golang.org/x/sync/errgroup.(*Group).Go.func1
         0     0% 93.89%      0.04s  1.11%  golang.org/x/sys/unix.SysctlUint64
         0     0% 93.89%      0.03s  0.83%  golang.org/x/sys/unix.nametomib
         0     0% 93.89%      0.04s  1.11%  golang.org/x/sys/unix.sysctl
         0     0% 93.89%      0.03s  0.83%  golang.org/x/sys/unix.sysctlmib
         0     0% 93.89%      0.15s  4.17%  internal/poll.(*FD).Write
         0     0% 93.89%      0.16s  4.44%  internal/poll.ignoringEINTRIO (inline)
         0     0% 93.89%      0.03s  0.83%  internal/runtime/maps.(*Map).growToSmall
         0     0% 93.89%      0.03s  0.83%  internal/runtime/maps.NewEmptyMap (inline)
         0     0% 93.89%      0.03s  0.83%  internal/runtime/maps.NewMap
         0     0% 93.89%      0.03s  0.83%  internal/runtime/maps.newGroups (inline)
         0     0% 93.89%      0.03s  0.83%  internal/runtime/maps.newarray
         0     0% 93.89%      0.09s  2.50%  io.Copy (inline)
         0     0% 93.89%      0.09s  2.50%  io.copyBuffer
         0     0% 93.89%      0.15s  4.17%  net.(*conn).Write
         0     0% 93.89%      0.15s  4.17%  net.(*netFD).Write
         0     0% 93.89%      0.23s  6.39%  net/http.(*Client).Do (inline)
         0     0% 93.89%      0.23s  6.39%  net/http.(*Client).do
         0     0% 93.89%      0.22s  6.11%  net/http.(*Client).send
         0     0% 93.89%      0.06s  1.67%  net/http.(*Request).write
         0     0% 93.89%      0.20s  5.56%  net/http.(*Transport).RoundTrip
         0     0% 93.89%      0.09s  2.50%  net/http.(*Transport).getConn
         0     0% 93.89%      0.02s  0.56%  net/http.(*Transport).prepareTransportCancel
         0     0% 93.89%      0.20s  5.56%  net/http.(*Transport).roundTrip
         0     0% 93.89%      0.04s  1.11%  net/http.(*persistConn).readLoop
         0     0% 93.89%      0.05s  1.39%  net/http.(*persistConn).roundTrip
         0     0% 93.89%      0.21s  5.83%  net/http.(*persistConn).writeLoop
         0     0% 93.89%      0.03s  0.83%  net/http.Header.Set (inline)
         0     0% 93.89%      0.03s  0.83%  net/http.NewRequest (inline)
         0     0% 93.89%      0.03s  0.83%  net/http.NewRequestWithContext
         0     0% 93.89%      0.15s  4.17%  net/http.persistConnWriter.Write
         0     0% 93.89%      0.22s  6.11%  net/http.send
         0     0% 93.89%      0.02s  0.56%  net/http.send.func1 (inline)
         0     0% 93.89%      0.03s  0.83%  net/textproto.MIMEHeader.Set (inline)
         0     0% 93.89%      0.02s  0.56%  net/url.(*URL).String
         0     0% 93.89%      0.03s  0.83%  net/url.Parse
         0     0% 93.89%      0.03s  0.83%  net/url.parse
         0     0% 93.89%      0.02s  0.56%  reflect.unsafe_New
         0     0% 93.89%      0.08s  2.22%  runtime.(*gcControllerState).enlistWorker
         0     0% 93.89%      0.08s  2.22%  runtime.(*gcWork).balance
         0     0% 93.89%      0.05s  1.39%  runtime.(*inlineUnwinder).next
         0     0% 93.89%      0.30s  8.33%  runtime.(*inlineUnwinder).resolveInternal (inline)
         0     0% 93.89%      0.02s  0.56%  runtime.(*mcache).nextFree
         0     0% 93.89%      0.02s  0.56%  runtime.(*mcache).prepareForSweep
         0     0% 93.89%      0.02s  0.56%  runtime.(*mcache).refill
         0     0% 93.89%      0.02s  0.56%  runtime.(*mcache).releaseAll
         0     0% 93.89%      0.02s  0.56%  runtime.(*mcentral).cacheSpan
         0     0% 93.89%      0.02s  0.56%  runtime.(*mcentral).uncacheSpan
         0     0% 93.89%      0.27s  7.50%  runtime.(*mheap).alloc.func1
         0     0% 93.89%      0.03s  0.83%  runtime.(*mheap).allocManual
         0     0% 93.89%      0.30s  8.33%  runtime.(*mheap).allocSpan
         0     0% 93.89%      0.05s  1.39%  runtime.(*pageAlloc).scavenge.func1
         0     0% 93.89%      0.05s  1.39%  runtime.(*pageAlloc).scavengeOne
         0     0% 93.89%      0.06s  1.67%  runtime.(*sweepLocked).sweep
         0     0% 93.89%      0.02s  0.56%  runtime.(*sweepLocked).sweep.(*mheap).freeSpan.func3
         0     0% 93.89%      0.15s  4.17%  runtime.(*unwinder).next
         0     0% 93.89%      0.02s  0.56%  runtime.acquirep
         0     0% 93.89%      0.04s  1.11%  runtime.addspecial
         0     0% 93.89%      0.04s  1.11%  runtime.bgsweep
         0     0% 93.89%      0.50s 13.89%  runtime.callers.func1
         0     0% 93.89%      0.02s  0.56%  runtime.cheaprandn (inline)
         0     0% 93.89%      0.02s  0.56%  runtime.concatstrings
         0     0% 93.89%      0.04s  1.11%  runtime.convT
         0     0% 93.89%      0.04s  1.11%  runtime.convTstring
         0     0% 93.89%      0.04s  1.11%  runtime.copystack
         0     0% 93.89%      0.70s 19.44%  runtime.findRunnable
         0     0% 93.89%      0.09s  2.50%  runtime.forEachPInternal
         0     0% 93.89%      0.06s  1.67%  runtime.freeSpecial
         0     0% 93.89%      0.10s  2.78%  runtime.funcspdelta (inline)
         0     0% 93.89%      0.04s  1.11%  runtime.gcAssistAlloc.func2
         0     0% 93.89%      0.04s  1.11%  runtime.gcAssistAlloc1
         0     0% 93.89%      0.21s  5.83%  runtime.gcBgMarkWorker
         0     0% 93.89%      0.32s  8.89%  runtime.gcBgMarkWorker.func2
         0     0% 93.89%      0.32s  8.89%  runtime.gcDrain
         0     0% 93.89%      0.30s  8.33%  runtime.gcDrainMarkWorkerDedicated (inline)
         0     0% 93.89%      0.02s  0.56%  runtime.gcDrainMarkWorkerIdle (inline)
         0     0% 93.89%      0.04s  1.11%  runtime.gcDrainN
         0     0% 93.89%      0.02s  0.56%  runtime.gcFlushBgCredit
         0     0% 93.89%      0.05s  1.39%  runtime.gcMarkDone.forEachP.func5
         0     0% 93.89%      0.04s  1.11%  runtime.gcMarkTermination.forEachP.func6
         0     0% 93.89%      0.02s  0.56%  runtime.gcMarkTermination.func3
         0     0% 93.89%      0.07s  1.94%  runtime.gcStart.func2
         0     0% 93.89%      0.05s  1.39%  runtime.gcstopm
         0     0% 93.89%      0.18s  5.00%  runtime.goexit0
         0     0% 93.89%      0.16s  4.44%  runtime.gopreempt_m (inline)
         0     0% 93.89%      0.16s  4.44%  runtime.goschedImpl
         0     0% 93.89%      0.05s  1.39%  runtime.growslice
         0     0% 93.89%      0.42s 11.67%  runtime.lock (partial-inline)
         0     0% 93.89%      0.42s 11.67%  runtime.lockWithRank (inline)
         0     0% 93.89%      0.30s  8.33%  runtime.mPark (inline)
         0     0% 93.89%      0.23s  6.39%  runtime.mProf_Malloc.func1
         0     0% 93.89%      0.03s  0.83%  runtime.makechan
         0     0% 93.89%      0.03s  0.83%  runtime.makemap
         0     0% 93.89%      0.03s  0.83%  runtime.makemap_small
         0     0% 93.89%      0.12s  3.33%  runtime.makeslice
         0     0% 93.89%      0.73s 20.28%  runtime.mallocgc
         0     0% 93.89%      0.05s  1.39%  runtime.mallocgcLarge
         0     0% 93.89%      0.14s  3.89%  runtime.mallocgcSmallNoscan
         0     0% 93.89%      0.47s 13.06%  runtime.mallocgcSmallScanNoHeader
         0     0% 93.89%      0.05s  1.39%  runtime.mallocgcTiny
         0     0% 93.89%      0.02s  0.56%  runtime.mapassign
         0     0% 93.89%      0.02s  0.56%  runtime.mapassign_faststr
         0     0% 93.89%      0.07s  1.94%  runtime.markroot.func1
         0     0% 93.89%      0.02s  0.56%  runtime.markrootBlock
         0     0% 93.89%      0.83s 23.06%  runtime.mcall
         0     0% 93.89%      0.03s  0.83%  runtime.memclrNoHeapPointersChunked
         0     0% 93.89%      0.17s  4.72%  runtime.morestack
         0     0% 93.89%      0.35s  9.72%  runtime.netpoll
         0     0% 93.89%      0.25s  6.94%  runtime.newInlineUnwinder
         0     0% 93.89%      0.04s  1.11%  runtime.newarray
         0     0% 93.89%      0.30s  8.33%  runtime.notesleep
         0     0% 93.89%      0.12s  3.33%  runtime.notetsleep
         0     0% 93.89%      0.12s  3.33%  runtime.notetsleep_internal
         0     0% 93.89%      0.47s 13.06%  runtime.notewakeup
         0     0% 93.89%      0.30s  8.33%  runtime.osyield (inline)
         0     0% 93.89%      0.64s 17.78%  runtime.park_m
         0     0% 93.89%      0.21s  5.83%  runtime.preemptM
         0     0% 93.89%      0.10s  2.78%  runtime.preemptall
         0     0% 93.89%      0.18s  5.00%  runtime.preemptone
         0     0% 93.89%      0.66s 18.33%  runtime.profilealloc
         0     0% 93.89%      0.02s  0.56%  runtime.rawstring (inline)
         0     0% 93.89%      0.02s  0.56%  runtime.rawstringtmp
         0     0% 93.89%      0.26s  7.22%  runtime.ready
         0     0% 93.89%      0.07s  1.94%  runtime.readyWithTime.goready.func1
         0     0% 93.89%      0.14s  3.89%  runtime.resetspinning
         0     0% 93.89%      0.05s  1.39%  runtime.runqgrab
         0     0% 93.89%      0.05s  1.39%  runtime.runqsteal
         0     0% 93.89%      0.04s  1.11%  runtime.scanstack
         0     0% 93.89%      0.86s 23.89%  runtime.schedule
         0     0% 93.89%      0.53s 14.72%  runtime.semasleep
         0     0% 93.89%      0.54s 15.00%  runtime.semawakeup
         0     0% 93.89%      0.17s  4.72%  runtime.send.goready.func1
         0     0% 93.89%      0.21s  5.83%  runtime.signalM (inline)
         0     0% 93.89%      0.02s  0.56%  runtime.slicebytetostring
         0     0% 93.89%      0.04s  1.11%  runtime.stackalloc
         0     0% 93.89%      0.02s  0.56%  runtime.stackcacherefill
         0     0% 93.89%      0.05s  1.39%  runtime.startTheWorld.func1
         0     0% 93.89%      0.08s  2.22%  runtime.startTheWorldWithSema
         0     0% 93.89%      0.41s 11.39%  runtime.startm
         0     0% 93.89%      0.05s  1.39%  runtime.stealWork
         0     0% 93.89%      0.07s  1.94%  runtime.stopTheWorld.func1
         0     0% 93.89%      0.15s  4.17%  runtime.stopTheWorldWithSema
         0     0% 93.89%      0.02s  0.56%  runtime.stoplockedm
         0     0% 93.89%      0.30s  8.33%  runtime.stopm
         0     0% 93.89%      0.03s  0.83%  runtime.suspendG
         0     0% 93.89%      0.04s  1.11%  runtime.sweepone
         0     0% 93.89%      0.05s  1.39%  runtime.sysUnused (inline)
         0     0% 93.89%      0.05s  1.39%  runtime.sysUnusedOS (inline)
         0     0% 93.89%      0.27s  7.50%  runtime.sysUsed (inline)
         0     0% 93.89%      0.27s  7.50%  runtime.sysUsedOS (inline)
         0     0% 93.89%      0.10s  2.78%  runtime.unlock (inline)
         0     0% 93.89%      0.07s  1.94%  runtime.unlock2Wake
         0     0% 93.89%      0.10s  2.78%  runtime.unlockWithRank (inline)
         0     0% 93.89%      0.41s 11.39%  runtime.wakep
         0     0% 93.89%      0.02s  0.56%  strings.(*Builder).Grow
         0     0% 93.89%      0.02s  0.56%  sync.(*Pool).Get
         0     0% 93.89%      0.15s  4.17%  syscall.Write (inline)
         0     0% 93.89%      0.15s  4.17%  syscall.write
```

### Memory (alloc_space)

Top: `go tool pprof -top -sample_index=alloc_space agent-mem.pprof`

```
File: agent-3011810301
Type: alloc_space
Time: 2026-04-16 11:19:40 MSK
Showing nodes accounting for 1491113.67kB, 96.92% of 1538536.52kB total
Dropped 511 nodes (cum <= 7692.68kB)
      flat  flat%   sum%        cum   cum%
772416.42kB 50.20% 50.20% 1396875.70kB 90.79%  compress/flate.NewWriter (inline)
381440.19kB 24.79% 75.00% 624459.28kB 40.59%  compress/flate.(*compressor).init
  238400kB 15.50% 90.49%   238400kB 15.50%  compress/flate.newDeflateFast (inline)
   38144kB  2.48% 92.97%    38144kB  2.48%  compress/flate.(*dictDecoder).init (inline)
   21456kB  1.39% 94.37%    21456kB  1.39%  regexp.(*bitState).reset
   19648kB  1.28% 95.64%    19648kB  1.28%  net/http.init.func15
 8977.25kB  0.58% 96.23% 47121.25kB  3.06%  compress/flate.NewReader
 7927.27kB  0.52% 96.74%  7927.27kB  0.52%  github.com/bq2cd/yp-go-metrics/internal/model.MetricSet.Upsert (inline)
  832.25kB 0.054% 96.80% 53652.89kB  3.49%  io.ReadAll
  819.50kB 0.053% 96.85% 52820.55kB  3.43%  compress/gzip.NewReader (inline)
  763.62kB  0.05% 96.90% 1501133.66kB 97.57%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).SendBatch
   93.17kB 0.0061% 96.90% 80545.99kB  5.24%  github.com/go-resty/resty/v2.(*Client).execute
   74.55kB 0.0048% 96.91% 21481.48kB  1.40%  net/http.(*Request).write
   55.88kB 0.0036% 96.91% 1405126.59kB 91.33%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).setBody
   55.88kB 0.0036% 96.92% 1489588.26kB 96.82%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.*uint8]).Do
    7.59kB 0.00049% 96.92% 1506553.33kB 97.92%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).reportWorker (inline)
    1.97kB 0.00013% 96.92% 21483.45kB  1.40%  net/http.(*persistConn).writeLoop
    0.05kB 3e-06% 96.92% 52001.05kB  3.38%  compress/gzip.(*Reader).Reset
    0.05kB 3e-06% 96.92% 47121.30kB  3.06%  compress/gzip.(*Reader).readHeader
    0.05kB 3e-06% 96.92% 1396950.30kB 90.80%  io.copyBuffer
         0     0% 96.92% 1396950.20kB 90.80%  bytes.(*Reader).WriteTo
         0     0% 96.92%  7745.33kB   0.5%  compress/flate.(*Writer).Close (inline)
         0     0% 96.92%  7745.33kB   0.5%  compress/flate.(*compressor).close
         0     0% 96.92%  7915.64kB  0.51%  compress/gzip.(*Writer).Close
         0     0% 96.92% 1396950.20kB 90.80%  compress/gzip.(*Writer).Write
         0     0% 96.92% 1506553.33kB 97.92%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).processBatches.func1
         0     0% 96.92% 1506545.73kB 97.92%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).reportBatch
         0     0% 96.92% 1487776.76kB 96.70%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendSingleRequest
         0     0% 96.92% 1492329.35kB 97.00%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendWithRetries
         0     0% 96.92% 1487776.76kB 96.70%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendWithRetries.func1
         0     0% 96.92% 15985.18kB  1.04%  github.com/go-resty/resty/v2.(*Client).executeBefore
         0     0% 96.92% 80545.99kB  5.24%  github.com/go-resty/resty/v2.(*Request).Execute
         0     0% 96.92% 80545.99kB  5.24%  github.com/go-resty/resty/v2.(*Request).Send (inline)
         0     0% 96.92% 21960.45kB  1.43%  github.com/go-resty/resty/v2.IsJSONType (inline)
         0     0% 96.92% 14504.24kB  0.94%  github.com/go-resty/resty/v2.parseRequestHeader
         0     0% 96.92% 53652.89kB  3.49%  github.com/go-resty/resty/v2.readAllWithLimit
         0     0% 96.92% 1511567.61kB 98.25%  golang.org/x/sync/errgroup.(*Group).Go.func1
         0     0% 96.92% 1396950.20kB 90.80%  io.Copy (inline)
         0     0% 96.92% 52820.64kB  3.43%  net/http.(*cancelTimerBody).Read
         0     0% 96.92% 52820.64kB  3.43%  net/http.(*gzipReader).Read
         0     0% 96.92% 20015.45kB  1.30%  net/http.(*transferWriter).doBodyCopy
         0     0% 96.92% 20043.44kB  1.30%  net/http.(*transferWriter).writeBody
         0     0% 96.92% 20011.59kB  1.30%  net/http.getCopyBuf (inline)
         0     0% 96.92% 21960.45kB  1.43%  regexp.(*Regexp).MatchString (inline)
         0     0% 96.92% 21960.45kB  1.43%  regexp.(*Regexp).backtrack
         0     0% 96.92% 21960.45kB  1.43%  regexp.(*Regexp).doExecute
         0     0% 96.92% 21960.45kB  1.43%  regexp.(*Regexp).doMatch (inline)
         0     0% 96.92% 27579.20kB  1.79%  sync.(*Pool).Get
```

## Server

### CPU

Top: `go tool pprof -top server-cpu.pprof`

```
File: server-168151938
Type: cpu
Time: 2026-04-16 11:14:40 MSK
Duration: 299.35s, Total samples = 8.09s ( 2.70%)
Showing nodes accounting for 7.45s, 92.09% of 8.09s total
Dropped 353 nodes (cum <= 0.04s)
      flat  flat%   sum%        cum   cum%
     1.85s 22.87% 22.87%      1.85s 22.87%  runtime.pthread_cond_signal
     1.01s 12.48% 35.35%      1.01s 12.48%  syscall.syscall
     0.95s 11.74% 47.10%      0.95s 11.74%  runtime.usleep
     0.78s  9.64% 56.74%      0.78s  9.64%  runtime.pthread_cond_wait
     0.70s  8.65% 65.39%      0.70s  8.65%  runtime.kevent
     0.43s  5.32% 70.70%      0.43s  5.32%  runtime.madvise
     0.34s  4.20% 74.91%      0.34s  4.20%  runtime.pthread_kill
     0.28s  3.46% 78.37%      0.68s  8.41%  runtime.pcvalue
     0.23s  2.84% 81.21%      0.27s  3.34%  runtime.step
     0.09s  1.11% 82.32%      0.09s  1.11%  runtime.(*moduledata).textAddr
     0.09s  1.11% 83.44%      1.01s 12.48%  runtime.tracebackPCs
     0.08s  0.99% 84.43%      0.09s  1.11%  runtime.typePointers.next
     0.07s  0.87% 85.29%      0.41s  5.07%  runtime.(*unwinder).resolveInternal
     0.07s  0.87% 86.16%      0.07s  0.87%  syscall.rawSyscall
     0.05s  0.62% 86.77%      0.06s  0.74%  runtime.(*unwinder).symPC
     0.05s  0.62% 87.39%      0.08s  0.99%  runtime.greyobject
     0.05s  0.62% 88.01%      0.19s  2.35%  runtime.scanobject
     0.04s  0.49% 88.50%      0.06s  0.74%  runtime.findfunc
     0.03s  0.37% 88.88%      0.47s  5.81%  runtime.(*unwinder).next
     0.03s  0.37% 89.25%      0.12s  1.48%  runtime.funcInfo.entry (inline)
     0.03s  0.37% 89.62%      4.54s 56.12%  runtime.systemstack
     0.03s  0.37% 89.99%      0.19s  2.35%  runtime.unlock2
     0.02s  0.25% 90.23%      0.21s  2.60%  runtime.(*sweepLocked).sweep
     0.02s  0.25% 90.48%      0.58s  7.17%  runtime.gcDrain
     0.02s  0.25% 90.73%      0.93s 11.50%  runtime.lock2
     0.02s  0.25% 90.98%      0.36s  4.45%  runtime.newInlineUnwinder
     0.02s  0.25% 91.22%      0.06s  0.74%  runtime.stkbucket
     0.01s  0.12% 91.35%      0.36s  4.45%  runtime.growslice
     0.01s  0.12% 91.47%      1.43s 17.68%  runtime.mallocgc
     0.01s  0.12% 91.59%      0.65s  8.03%  runtime.netpoll
     0.01s  0.12% 91.72%      0.34s  4.20%  runtime.pcdatavalue1
     0.01s  0.12% 91.84%      0.05s  0.62%  runtime.scanblock
     0.01s  0.12% 91.97%      0.08s  0.99%  runtime.scanstack
     0.01s  0.12% 92.09%      0.16s  1.98%  runtime.unlock2Wake
         0     0% 92.09%      0.05s  0.62%  bufio.(*Reader).Peek
         0     0% 92.09%      0.05s  0.62%  bufio.(*Reader).fill
         0     0% 92.09%      0.14s  1.73%  bufio.(*Writer).Flush
         0     0% 92.09%      0.12s  1.48%  database/sql.(*DB).BeginTx
         0     0% 92.09%      0.12s  1.48%  database/sql.(*DB).BeginTx.func1
         0     0% 92.09%      0.52s  6.43%  database/sql.(*DB).QueryContext
         0     0% 92.09%      0.52s  6.43%  database/sql.(*DB).QueryContext.func1
         0     0% 92.09%      0.12s  1.48%  database/sql.(*DB).begin
         0     0% 92.09%      0.12s  1.48%  database/sql.(*DB).beginDC
         0     0% 92.09%      0.12s  1.48%  database/sql.(*DB).beginDC.func1
         0     0% 92.09%      0.29s  3.58%  database/sql.(*DB).conn
         0     0% 92.09%      0.08s  0.99%  database/sql.(*DB).execDC
         0     0% 92.09%      0.08s  0.99%  database/sql.(*DB).execDC.func2
         0     0% 92.09%      0.52s  6.43%  database/sql.(*DB).query
         0     0% 92.09%      0.22s  2.72%  database/sql.(*DB).queryDC
         0     0% 92.09%      0.21s  2.60%  database/sql.(*DB).queryDC.func1
         0     0% 92.09%      0.64s  7.91%  database/sql.(*DB).retry
         0     0% 92.09%      0.11s  1.36%  database/sql.(*Rows).Next
         0     0% 92.09%      0.11s  1.36%  database/sql.(*Rows).Next.func1
         0     0% 92.09%      0.11s  1.36%  database/sql.(*Rows).nextLocked
         0     0% 92.09%      0.08s  0.99%  database/sql.(*Tx).ExecContext
         0     0% 92.09%      0.12s  1.48%  database/sql.ctxDriverBegin
         0     0% 92.09%      0.08s  0.99%  database/sql.ctxDriverExec
         0     0% 92.09%      0.19s  2.35%  database/sql.ctxDriverQuery
         0     0% 92.09%      0.53s  6.55%  database/sql.withLock
         0     0% 92.09%      0.22s  2.72%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).createPeriodicTask.func2
         0     0% 92.09%      0.22s  2.72%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).dumpMetrics
         0     0% 92.09%      0.64s  7.91%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).launchBatchWriter.func1
         0     0% 92.09%      0.22s  2.72%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).launchMetricDumper.func1
         0     0% 92.09%      1.05s 12.98%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).ServeHTTP
         0     0% 92.09%      0.86s 10.63%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).retrieveMetrics
         0     0% 92.09%      0.06s  0.74%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).validateMetrics
         0     0% 92.09%      1.40s 17.31%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorMiddleware).Intercept
         0     0% 92.09%      1.12s 13.84%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).Intercept
         0     0% 92.09%      1.35s 16.69%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*loggerMiddleware).Intercept
         0     0% 92.09%      1.42s 17.55%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*middlewareHandler).ServeHTTP
         0     0% 92.09%      1.05s 12.98%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*recovererMiddleware).Intercept
         0     0% 92.09%      1.42s 17.55%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*requestIDMiddleware).Intercept
         0     0% 92.09%      1.43s 17.68%  github.com/bq2cd/yp-go-metrics/internal/handler/router.(*Router).ServeHTTP
         0     0% 92.09%      0.07s  0.87%  github.com/bq2cd/yp-go-metrics/internal/model.MetricKey.Compare (inline)
         0     0% 92.09%      0.05s  0.62%  github.com/bq2cd/yp-go-metrics/internal/model.NewAuditEvent
         0     0% 92.09%      0.05s  0.62%  github.com/bq2cd/yp-go-metrics/internal/model.NewAuditMetricName (inline)
         0     0% 92.09%      0.17s  2.10%  github.com/bq2cd/yp-go-metrics/internal/repository/auditsink.(*fileSink).WriteEvent
         0     0% 92.09%      0.05s  0.62%  github.com/bq2cd/yp-go-metrics/internal/repository/auditsink.NewFileSink.func1
         0     0% 92.09%      0.07s  0.87%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).GetAll
         0     0% 92.09%      0.97s 11.99%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).GetMulti
         0     0% 92.09%      0.48s  5.93%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).SetMulti
         0     0% 92.09%         1s 12.36%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByType
         0     0% 92.09%      1.03s 12.73%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByTypeWithRetries
         0     0% 92.09%         1s 12.36%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByTypeWithRetries.func1
         0     0% 92.09%      0.99s 12.24%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiForType
         0     0% 92.09%      0.45s  5.56%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMulti
         0     0% 92.09%      0.45s  5.56%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiForType
         0     0% 92.09%      0.48s  5.93%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiWithRetries
         0     0% 92.09%      0.45s  5.56%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiWithRetries.func1
         0     0% 92.09%      0.10s  1.24%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].ConvertMetrics
         0     0% 92.09%      0.07s  0.87%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].ConvertMetrics.func1
         0     0% 92.09%      0.20s  2.47%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].Insert
         0     0% 92.09%      0.76s  9.39%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].Select
         0     0% 92.09%      0.23s  2.84%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value int64 "db:\"value\"" }].Select
         0     0% 92.09%      0.17s  2.10%  github.com/bq2cd/yp-go-metrics/internal/service.(*auditEventProcessor).processEvent.func1
         0     0% 92.09%      0.05s  0.62%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricAuditor).RecordMetricsUploaded
         0     0% 92.09%      0.05s  0.62%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricJSONEncoder).EncodeBatch
         0     0% 92.09%      0.14s  1.73%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricSnapshotter).DumpClose
         0     0% 92.09%      0.07s  0.87%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricStorer).RetrieveAll
         0     0% 92.09%      0.86s 10.63%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricStorer).RetrieveBatch
         0     0% 92.09%      0.64s  7.91%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).StartProcessing
         0     0% 92.09%      0.16s  1.98%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).accumulateCounters
         0     0% 92.09%      0.64s  7.91%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).processBatchTx
         0     0% 92.09%      0.06s  0.74%  github.com/bq2cd/yp-go-metrics/pkg/log.(*baseLogger).With
         0     0% 92.09%      0.20s  2.47%  github.com/bq2cd/yp-go-metrics/pkg/log.(*eventBuilder).Msg
         0     0% 92.09%      0.20s  2.47%  github.com/bq2cd/yp-go-metrics/pkg/log.(*zapLogger).log
         0     0% 92.09%      0.22s  2.72%  github.com/bq2cd/yp-go-metrics/pkg/periodictask.(*timerTask).Run
         0     0% 92.09%      1.01s 12.48%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.[]github.com/bq2cd/yp-go-metrics/internal/model.Metric]).Do
         0     0% 92.09%      0.47s  5.81%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.interface {}]).Do
         0     0% 92.09%      1.43s 17.68%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 92.09%      1.05s 12.98%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 92.09%      0.06s  0.74%  github.com/goccy/go-json.(*Decoder).Decode (inline)
         0     0% 92.09%      0.06s  0.74%  github.com/goccy/go-json.(*Decoder).DecodeWithOption
         0     0% 92.09%      0.06s  0.74%  github.com/goccy/go-json.(*Encoder).Encode (inline)
         0     0% 92.09%      0.06s  0.74%  github.com/goccy/go-json.(*Encoder).EncodeWithOption
         0     0% 92.09%      0.05s  0.62%  github.com/goccy/go-json.(*Encoder).encodeWithOption
         0     0% 92.09%      0.05s  0.62%  github.com/goccy/go-json.MarshalContext (inline)
         0     0% 92.09%      0.05s  0.62%  github.com/goccy/go-json.marshalContext
         0     0% 92.09%      0.06s  0.74%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).DecodeStream
         0     0% 92.09%      0.19s  2.35%  github.com/huandu/go-sqlbuilder.(*Args).CompileWithFlavor
         0     0% 92.09%      0.17s  2.10%  github.com/huandu/go-sqlbuilder.(*Args).compileDigits
         0     0% 92.09%      0.17s  2.10%  github.com/huandu/go-sqlbuilder.(*Args).compileSuccessive
         0     0% 92.09%      0.10s  1.24%  github.com/huandu/go-sqlbuilder.(*Cond).In.func1
         0     0% 92.09%      0.08s  0.99%  github.com/huandu/go-sqlbuilder.(*InsertBuilder).BuildWithFlavor
         0     0% 92.09%      0.21s  2.60%  github.com/huandu/go-sqlbuilder.(*SelectBuilder).BuildWithFlavor
         0     0% 92.09%      0.15s  1.85%  github.com/huandu/go-sqlbuilder.(*WhereClause).BuildWithFlavor
         0     0% 92.09%      0.17s  2.10%  github.com/huandu/go-sqlbuilder.(*argsCompileContext).WriteValue
         0     0% 92.09%      0.08s  0.99%  github.com/huandu/go-sqlbuilder.(*argsCompileContext).WriteValues
         0     0% 92.09%      0.13s  1.61%  github.com/huandu/go-sqlbuilder.(*clause).Build
         0     0% 92.09%      0.11s  1.36%  github.com/huandu/go-sqlbuilder.(*stringBuilder).WriteString (inline)
         0     0% 92.09%      0.05s  0.62%  github.com/huandu/go-sqlbuilder.(*stringBuilder).WriteStrings
         0     0% 92.09%      0.12s  1.48%  github.com/jackc/pgx/v5.(*Conn).BeginTx
         0     0% 92.09%      0.21s  2.60%  github.com/jackc/pgx/v5.(*Conn).Exec
         0     0% 92.09%      0.18s  2.22%  github.com/jackc/pgx/v5.(*Conn).Query
         0     0% 92.09%      0.21s  2.60%  github.com/jackc/pgx/v5.(*Conn).exec
         0     0% 92.09%      0.08s  0.99%  github.com/jackc/pgx/v5.(*Conn).execPrepared
         0     0% 92.09%      0.13s  1.61%  github.com/jackc/pgx/v5.(*Conn).execSimpleProtocol
         0     0% 92.09%      0.22s  2.72%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 92.09%      0.07s  0.87%  github.com/jackc/pgx/v5.ParseConfig (inline)
         0     0% 92.09%      0.07s  0.87%  github.com/jackc/pgx/v5.ParseConfigWithOptions
         0     0% 92.09%      0.22s  2.72%  github.com/jackc/pgx/v5.connect
         0     0% 92.09%      0.12s  1.48%  github.com/jackc/pgx/v5/pgconn.(*PgConn).Exec
         0     0% 92.09%      0.24s  2.97%  github.com/jackc/pgx/v5/pgconn.(*PgConn).ExecPrepared
         0     0% 92.09%      0.07s  0.87%  github.com/jackc/pgx/v5/pgconn.(*PgConn).enterPotentialWriteReadDeadlock (inline)
         0     0% 92.09%      0.23s  2.84%  github.com/jackc/pgx/v5/pgconn.(*PgConn).execExtendedSuffix
         0     0% 92.09%      0.32s  3.96%  github.com/jackc/pgx/v5/pgconn.(*PgConn).flushWithPotentialWriteReadDeadlock
         0     0% 92.09%      0.05s  0.62%  github.com/jackc/pgx/v5/pgconn.(*PgConn).peekMessage
         0     0% 92.09%      0.06s  0.74%  github.com/jackc/pgx/v5/pgconn.(*ResultReader).readUntilRowDescription
         0     0% 92.09%      0.22s  2.72%  github.com/jackc/pgx/v5/pgconn.ConnectConfig
         0     0% 92.09%      0.07s  0.87%  github.com/jackc/pgx/v5/pgconn.ParseConfigWithOptions
         0     0% 92.09%      0.19s  2.35%  github.com/jackc/pgx/v5/pgconn.connectOne
         0     0% 92.09%      0.19s  2.35%  github.com/jackc/pgx/v5/pgconn.connectPreferred
         0     0% 92.09%      0.05s  0.62%  github.com/jackc/pgx/v5/pgconn.defaultSettings
         0     0% 92.09%      0.05s  0.62%  github.com/jackc/pgx/v5/pgconn/internal/bgreader.(*BGReader).Read
         0     0% 92.09%      0.25s  3.09%  github.com/jackc/pgx/v5/pgproto3.(*Frontend).Flush
         0     0% 92.09%      0.05s  0.62%  github.com/jackc/pgx/v5/pgproto3.(*Frontend).Receive
         0     0% 92.09%      0.05s  0.62%  github.com/jackc/pgx/v5/pgproto3.(*chunkReader).Next
         0     0% 92.09%      0.05s  0.62%  github.com/jackc/pgx/v5/pgtype.scanPlanString.Scan
         0     0% 92.09%      0.12s  1.48%  github.com/jackc/pgx/v5/stdlib.(*Conn).BeginTx
         0     0% 92.09%      0.08s  0.99%  github.com/jackc/pgx/v5/stdlib.(*Conn).ExecContext
         0     0% 92.09%      0.19s  2.35%  github.com/jackc/pgx/v5/stdlib.(*Conn).QueryContext
         0     0% 92.09%      0.11s  1.36%  github.com/jackc/pgx/v5/stdlib.(*Rows).Next
         0     0% 92.09%      0.08s  0.99%  github.com/jackc/pgx/v5/stdlib.(*Rows).Next.func14
         0     0% 92.09%      0.29s  3.58%  github.com/jackc/pgx/v5/stdlib.(*driverConnector).Connect
         0     0% 92.09%      0.12s  1.48%  github.com/jmoiron/sqlx.(*DB).BeginTxx
         0     0% 92.09%      0.52s  6.43%  github.com/jmoiron/sqlx.(*DB).QueryxContext
         0     0% 92.09%      0.70s  8.65%  github.com/jmoiron/sqlx.(*DB).SelectContext
         0     0% 92.09%      0.70s  8.65%  github.com/jmoiron/sqlx.SelectContext
         0     0% 92.09%      0.18s  2.22%  github.com/jmoiron/sqlx.scanAll
         0     0% 92.09%      0.19s  2.35%  go.uber.org/zap.(*Logger).Log
         0     0% 92.09%      0.05s  0.62%  go.uber.org/zap.(*Logger).check
         0     0% 92.09%      0.14s  1.73%  go.uber.org/zap/zapcore.(*CheckedEntry).Write
         0     0% 92.09%      0.14s  1.73%  go.uber.org/zap/zapcore.(*ioCore).Write
         0     0% 92.09%      0.13s  1.61%  go.uber.org/zap/zapcore.(*lockedWriteSyncer).Write
         0     0% 92.09%      0.17s  2.10%  golang.org/x/sync/errgroup.(*Group).Go.func1
         0     0% 92.09%      0.14s  1.73%  internal/poll.(*FD).Read
         0     0% 92.09%      0.68s  8.41%  internal/poll.(*FD).Write
         0     0% 92.09%      0.82s 10.14%  internal/poll.ignoringEINTRIO (inline)
         0     0% 92.09%      0.08s  0.99%  io.ReadAtLeast
         0     0% 92.09%      0.12s  1.48%  net.(*Dialer).DialContext
         0     0% 92.09%      0.14s  1.73%  net.(*conn).Read
         0     0% 92.09%      0.40s  4.94%  net.(*conn).Write
         0     0% 92.09%      0.14s  1.73%  net.(*netFD).Read
         0     0% 92.09%      0.40s  4.94%  net.(*netFD).Write
         0     0% 92.09%      0.12s  1.48%  net.(*sysDialer).dialParallel
         0     0% 92.09%      0.12s  1.48%  net.(*sysDialer).dialSerial
         0     0% 92.09%      0.12s  1.48%  net.(*sysDialer).dialSingle
         0     0% 92.09%      0.12s  1.48%  net.(*sysDialer).dialTCP
         0     0% 92.09%      0.12s  1.48%  net.(*sysDialer).doDialTCP (inline)
         0     0% 92.09%      0.12s  1.48%  net.(*sysDialer).doDialTCPProto
         0     0% 92.09%      0.10s  1.24%  net.internetSocket
         0     0% 92.09%      0.10s  1.24%  net.socket
         0     0% 92.09%      0.06s  0.74%  net.sysSocket
         0     0% 92.09%      1.64s 20.27%  net/http.(*conn).serve
         0     0% 92.09%      0.05s  0.62%  net/http.(*connReader).Read
         0     0% 92.09%      0.14s  1.73%  net/http.(*response).finishRequest
         0     0% 92.09%      1.05s 12.98%  net/http.HandlerFunc.ServeHTTP
         0     0% 92.09%      0.13s  1.61%  net/http.checkConnErrorWriter.Write
         0     0% 92.09%      1.43s 17.68%  net/http.serverHandler.ServeHTTP
         0     0% 92.09%      0.28s  3.46%  os.(*File).Write
         0     0% 92.09%      0.28s  3.46%  os.(*File).write (inline)
         0     0% 92.09%      0.05s  0.62%  os.OpenFile
         0     0% 92.09%      0.05s  0.62%  os.Stat
         0     0% 92.09%      0.13s  1.61%  os.ignoringEINTR (inline)
         0     0% 92.09%      0.05s  0.62%  os.open (inline)
         0     0% 92.09%      0.05s  0.62%  os.openFileNolog
         0     0% 92.09%      0.05s  0.62%  os.openFileNolog.func1 (inline)
         0     0% 92.09%      0.05s  0.62%  os.statNolog
         0     0% 92.09%      0.05s  0.62%  os.statNolog.func1 (inline)
         0     0% 92.09%      0.05s  0.62%  reflect.Append
         0     0% 92.09%      0.05s  0.62%  reflect.Value.extendSlice
         0     0% 92.09%      0.23s  2.84%  runtime.(*gcControllerState).enlistWorker
         0     0% 92.09%      0.23s  2.84%  runtime.(*gcWork).balance
         0     0% 92.09%      0.34s  4.20%  runtime.(*inlineUnwinder).resolveInternal (inline)
         0     0% 92.09%      0.17s  2.10%  runtime.(*mcache).prepareForSweep
         0     0% 92.09%      0.17s  2.10%  runtime.(*mcache).releaseAll
         0     0% 92.09%      0.17s  2.10%  runtime.(*mcentral).uncacheSpan
         0     0% 92.09%      0.32s  3.96%  runtime.(*mheap).alloc.func1
         0     0% 92.09%      0.07s  0.87%  runtime.(*mheap).allocManual
         0     0% 92.09%      0.39s  4.82%  runtime.(*mheap).allocSpan
         0     0% 92.09%      0.05s  0.62%  runtime.(*pageAlloc).scavenge.func1
         0     0% 92.09%      0.05s  0.62%  runtime.(*pageAlloc).scavengeOne
         0     0% 92.09%      0.07s  0.87%  runtime.(*timer).maybeAdd
         0     0% 92.09%      0.07s  0.87%  runtime.(*timer).modify
         0     0% 92.09%      0.07s  0.87%  runtime.(*timer).reset (inline)
         0     0% 92.09%      0.08s  0.99%  runtime.acquirep
         0     0% 92.09%      0.08s  0.99%  runtime.addspecial
         0     0% 92.09%      1.04s 12.86%  runtime.callers
         0     0% 92.09%      1.04s 12.86%  runtime.callers.func1
         0     0% 92.09%      0.09s  1.11%  runtime.convTstring
         0     0% 92.09%      0.08s  0.99%  runtime.copystack
         0     0% 92.09%      0.07s  0.87%  runtime.deductAssistCredit
         0     0% 92.09%      1.60s 19.78%  runtime.findRunnable
         0     0% 92.09%      0.18s  2.22%  runtime.forEachPInternal
         0     0% 92.09%      0.19s  2.35%  runtime.freeSpecial
         0     0% 92.09%      0.34s  4.20%  runtime.funcspdelta (inline)
         0     0% 92.09%      0.07s  0.87%  runtime.gcAssistAlloc
         0     0% 92.09%      0.07s  0.87%  runtime.gcAssistAlloc.func2
         0     0% 92.09%      0.07s  0.87%  runtime.gcAssistAlloc1
         0     0% 92.09%      0.32s  3.96%  runtime.gcBgMarkWorker
         0     0% 92.09%      0.58s  7.17%  runtime.gcBgMarkWorker.func2
         0     0% 92.09%      0.41s  5.07%  runtime.gcDrainMarkWorkerDedicated (inline)
         0     0% 92.09%      0.17s  2.10%  runtime.gcDrainMarkWorkerIdle (inline)
         0     0% 92.09%      0.07s  0.87%  runtime.gcDrainN
         0     0% 92.09%      0.07s  0.87%  runtime.gcMarkDone
         0     0% 92.09%      0.07s  0.87%  runtime.gcMarkTermination
         0     0% 92.09%      0.14s  1.73%  runtime.gcMarkTermination.forEachP.func6
         0     0% 92.09%      0.08s  0.99%  runtime.gcMarkTermination.func3
         0     0% 92.09%      0.08s  0.99%  runtime.gcMarkTermination.func4
         0     0% 92.09%      0.06s  0.74%  runtime.gcstopm
         0     0% 92.09%      0.21s  2.60%  runtime.goexit0
         0     0% 92.09%      0.27s  3.34%  runtime.gopreempt_m (inline)
         0     0% 92.09%      0.27s  3.34%  runtime.goschedImpl
         0     0% 92.09%      0.93s 11.50%  runtime.lock (inline)
         0     0% 92.09%      0.93s 11.50%  runtime.lockWithRank (inline)
         0     0% 92.09%      0.64s  7.91%  runtime.mPark (inline)
         0     0% 92.09%      0.06s  0.74%  runtime.mProf_Free
         0     0% 92.09%      1.30s 16.07%  runtime.mProf_Malloc
         0     0% 92.09%      0.64s  7.91%  runtime.mProf_Malloc.func1
         0     0% 92.09%      0.21s  2.60%  runtime.makeslice
         0     0% 92.09%      0.18s  2.22%  runtime.mallocgcSmallNoscan
         0     0% 92.09%      0.94s 11.62%  runtime.mallocgcSmallScanNoHeader
         0     0% 92.09%      0.19s  2.35%  runtime.mallocgcTiny
         0     0% 92.09%      0.19s  2.35%  runtime.markroot
         0     0% 92.09%      0.15s  1.85%  runtime.markroot.func1
         0     0% 92.09%      1.68s 20.77%  runtime.mcall
         0     0% 92.09%      0.32s  3.96%  runtime.morestack
         0     0% 92.09%      0.06s  0.74%  runtime.netpollBreak (inline)
         0     0% 92.09%      0.05s  0.62%  runtime.newarray
         0     0% 92.09%      0.52s  6.43%  runtime.newobject
         0     0% 92.09%      0.96s 11.87%  runtime.newproc.func1
         0     0% 92.09%      0.34s  4.20%  runtime.newstack
         0     0% 92.09%      0.64s  7.91%  runtime.notesleep
         0     0% 92.09%      1.72s 21.26%  runtime.notewakeup
         0     0% 92.09%      0.77s  9.52%  runtime.osyield (inline)
         0     0% 92.09%      1.45s 17.92%  runtime.park_m
         0     0% 92.09%      0.34s  4.20%  runtime.preemptM
         0     0% 92.09%      0.06s  0.74%  runtime.preemptall
         0     0% 92.09%      0.29s  3.58%  runtime.preemptone
         0     0% 92.09%      1.30s 16.07%  runtime.profilealloc
         0     0% 92.09%      0.52s  6.43%  runtime.ready
         0     0% 92.09%      0.18s  2.22%  runtime.resetspinning
         0     0% 92.09%      0.18s  2.22%  runtime.runqgrab
         0     0% 92.09%      0.18s  2.22%  runtime.runqsteal
         0     0% 92.09%      1.81s 22.37%  runtime.schedule
         0     0% 92.09%      0.81s 10.01%  runtime.semasleep
         0     0% 92.09%      1.87s 23.11%  runtime.semawakeup
         0     0% 92.09%      0.49s  6.06%  runtime.send.goready.func1
         0     0% 92.09%      0.64s  7.91%  runtime.setprofilebucket
         0     0% 92.09%      0.34s  4.20%  runtime.signalM (inline)
         0     0% 92.09%      0.10s  1.24%  runtime.slicebytetostring
         0     0% 92.09%      0.05s  0.62%  runtime.stackalloc
         0     0% 92.09%      0.09s  1.11%  runtime.startTheWorldWithSema
         0     0% 92.09%      1.68s 20.77%  runtime.startm
         0     0% 92.09%      0.18s  2.22%  runtime.stealWork
         0     0% 92.09%      0.71s  8.78%  runtime.stopm
         0     0% 92.09%      0.07s  0.87%  runtime.suspendG
         0     0% 92.09%      0.05s  0.62%  runtime.sysUnused (inline)
         0     0% 92.09%      0.05s  0.62%  runtime.sysUnusedOS (inline)
         0     0% 92.09%      0.38s  4.70%  runtime.sysUsed (inline)
         0     0% 92.09%      0.38s  4.70%  runtime.sysUsedOS (inline)
         0     0% 92.09%      0.19s  2.35%  runtime.unlock (inline)
         0     0% 92.09%      0.19s  2.35%  runtime.unlockWithRank (inline)
         0     0% 92.09%      0.06s  0.74%  runtime.wakeNetPoller
         0     0% 92.09%      0.06s  0.74%  runtime.wakeNetpoll
         0     0% 92.09%      1.73s 21.38%  runtime.wakep
         0     0% 92.09%      0.07s  0.87%  slices.SortStableFunc[go.shape.[]go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" },go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }] (inline)
         0     0% 92.09%      0.09s  1.11%  slices.SortedStableFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }]
         0     0% 92.09%      0.07s  0.87%  slices.insertionSortCmpFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }] (inline)
         0     0% 92.09%      0.07s  0.87%  slices.stableCmpFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }]
         0     0% 92.09%      0.14s  1.73%  strings.(*Builder).WriteString (inline)
         0     0% 92.09%      0.11s  1.36%  sync.(*Pool).Get
         0     0% 92.09%      0.06s  0.74%  sync.(*Pool).pin
         0     0% 92.09%      0.06s  0.74%  sync.(*Pool).pinSlow
         0     0% 92.09%      0.05s  0.62%  syscall.Open
         0     0% 92.09%      0.14s  1.73%  syscall.Read (inline)
         0     0% 92.09%      0.06s  0.74%  syscall.Socket
         0     0% 92.09%      0.05s  0.62%  syscall.Stat
         0     0% 92.09%      0.68s  8.41%  syscall.Write (inline)
         0     0% 92.09%      0.14s  1.73%  syscall.read
         0     0% 92.09%      0.06s  0.74%  syscall.socket
         0     0% 92.09%      0.68s  8.41%  syscall.write
         0     0% 92.09%      0.07s  0.87%  time.(*Timer).Reset
         0     0% 92.09%      0.07s  0.87%  time.resetTimer
```

### Memory (alloc_space)

Top: `go tool pprof -top -sample_index=alloc_space server-mem.pprof`

```
File: server-168151938
Type: alloc_space
Time: 2026-04-16 11:19:40 MSK
Showing nodes accounting for 1449510.84kB, 93.46% of 1550966.55kB total
Dropped 1000 nodes (cum <= 7754.83kB)
      flat  flat%   sum%        cum   cum%
  770472kB 49.68% 49.68% 1393359.38kB 89.84%  compress/flate.NewWriter (inline)
  380480kB 24.53% 74.21% 622887.38kB 40.16%  compress/flate.(*compressor).init
  237800kB 15.33% 89.54%   237800kB 15.33%  compress/flate.newDeflateFast (inline)
   38144kB  2.46% 92.00%    38144kB  2.46%  compress/flate.(*dictDecoder).init (inline)
 8977.25kB  0.58% 92.58% 47121.30kB  3.04%  compress/flate.NewReader
 8192.28kB  0.53% 93.11%  8192.28kB  0.53%  github.com/pressly/goose/v3/internal/sqlparser.init.func1
 1103.16kB 0.071% 93.18% 20159.27kB  1.30%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).GetMulti
  927.12kB  0.06% 93.24% 14643.81kB  0.94%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiForType
  819.50kB 0.053% 93.29% 52839.31kB  3.41%  compress/gzip.NewReader (inline)
  805.44kB 0.052% 93.34% 15586.64kB  1.00%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByType
  789.82kB 0.051% 93.39%  8694.48kB  0.56%  github.com/bq2cd/yp-go-metrics/pkg/log.(*zapLogger).with
  538.73kB 0.035% 93.43% 11633.53kB  0.75%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].Select
  186.25kB 0.012% 93.44% 1435391.52kB 92.55%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*loggerMiddleware).Intercept
  104.39kB 0.0067% 93.45%  8798.88kB  0.57%  github.com/bq2cd/yp-go-metrics/pkg/log.(*baseLogger).With
   72.47kB 0.0047% 93.45% 17777.20kB  1.15%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.[]github.com/bq2cd/yp-go-metrics/internal/model.Metric]).Do
   55.83kB 0.0036% 93.46% 24812.95kB  1.60%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.interface {}]).Do
   18.81kB 0.0012% 93.46% 1510081.81kB 97.36%  net/http.(*conn).serve
   18.62kB 0.0012% 93.46% 1496484.23kB 96.49%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*requestIDMiddleware).Intercept
    4.31kB 0.00028% 93.46% 25150.24kB  1.62%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).StartProcessing
    0.27kB 1.7e-05% 93.46%  8197.18kB  0.53%  github.com/pressly/goose/v3/internal/sqlparser.ParseSQLMigration
    0.12kB 8.1e-06% 93.46%  8244.74kB  0.53%  github.com/pressly/goose/v3.(*Migration).run
    0.09kB 6e-06% 93.46% 47140.02kB  3.04%  compress/gzip.(*Reader).readHeader
    0.09kB 6e-06% 93.46%  7816.34kB   0.5%  io.copyBuffer
    0.06kB 4e-06% 93.46%  8446.91kB  0.54%  github.com/bq2cd/yp-go-metrics/internal/app/server.applyMigrations
    0.05kB 3e-06% 93.46% 52019.81kB  3.35%  compress/gzip.(*Reader).Reset
    0.05kB 3e-06% 93.46% 52839.36kB  3.41%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorMiddleware).decompressRequest
    0.05kB 3e-06% 93.46%  8077.70kB  0.52%  net/http.(*chunkWriter).writeHeader
    0.04kB 2.5e-06% 93.46%  8438.73kB  0.54%  github.com/pressly/goose/v3.UpToContext
    0.03kB 2e-06% 93.46%  9577.02kB  0.62%  github.com/bq2cd/yp-go-metrics/internal/app/cli.App[go.shape.struct { ListenAddress string; ShutdownTimeout time.Duration; MetricStoreInterval time.Duration; MetricStoreFilePath string; MetricStoreLoadOnStartup bool; DatabaseURL net/url.URL; HMACSecretKey []uint8; AuditFilePath string; AuditURL net/url.URL }].Run
         0     0% 93.46%  8077.70kB  0.52%  bufio.(*Writer).Flush
         0     0% 93.46% 1393359.38kB 89.84%  compress/gzip.(*Writer).Write
         0     0% 93.46%  8518.41kB  0.55%  database/sql.(*DB).retry
         0     0% 93.46%  8552.11kB  0.55%  github.com/bq2cd/yp-go-metrics/internal/app/cli.App[go.shape.struct { ListenAddress string; ShutdownTimeout time.Duration; MetricStoreInterval time.Duration; MetricStoreFilePath string; MetricStoreLoadOnStartup bool; DatabaseURL net/url.URL; HMACSecretKey []uint8; AuditFilePath string; AuditURL net/url.URL }].run
         0     0% 93.46% 25150.24kB  1.62%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).launchBatchWriter.func1
         0     0% 93.46%  8552.11kB  0.55%  github.com/bq2cd/yp-go-metrics/internal/app/server.Run
         0     0% 93.46%  8451.21kB  0.54%  github.com/bq2cd/yp-go-metrics/internal/app/server.applyMigrationsWithRetries
         0     0% 93.46%  8446.91kB  0.54%  github.com/bq2cd/yp-go-metrics/internal/app/server.applyMigrationsWithRetries.func1
         0     0% 93.46%  8481.99kB  0.55%  github.com/bq2cd/yp-go-metrics/internal/app/server.initStorage
         0     0% 93.46% 30757.63kB  1.98%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).ServeHTTP
         0     0% 93.46% 19248.77kB  1.24%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).retrieveMetrics
         0     0% 93.46% 1495943.93kB 96.45%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorMiddleware).Intercept
         0     0% 93.46% 1393359.38kB 89.84%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorResponseWriter).Write
         0     0% 93.46% 1427937.15kB 92.07%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).Intercept
         0     0% 93.46% 1394678.42kB 89.92%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).writeResponse
         0     0% 93.46% 1393359.38kB 89.84%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*loggerResponseWriter).Write
         0     0% 93.46% 1496484.23kB 96.49%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*middlewareHandler).ServeHTTP
         0     0% 93.46% 30766.26kB  1.98%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*recovererMiddleware).Intercept
         0     0% 93.46% 1497664.87kB 96.56%  github.com/bq2cd/yp-go-metrics/internal/handler/router.(*Router).ServeHTTP
         0     0% 93.46% 18123.80kB  1.17%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).SetMulti
         0     0% 93.46% 20091.60kB  1.30%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByTypeWithRetries
         0     0% 93.46% 15586.64kB  1.00%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByTypeWithRetries.func1
         0     0% 93.46% 14657.08kB  0.95%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMulti
         0     0% 93.46% 11582.80kB  0.75%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiForType
         0     0% 93.46% 18123.80kB  1.17%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiWithRetries
         0     0% 93.46% 14657.08kB  0.95%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiWithRetries.func1
         0     0% 93.46% 19248.77kB  1.24%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricStorer).RetrieveBatch
         0     0% 93.46% 25145.93kB  1.62%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).processBatchTx
         0     0% 93.46% 1497664.87kB 96.56%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 93.46% 30766.26kB  1.98%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 93.46%  9728.11kB  0.63%  github.com/jmoiron/sqlx.(*DB).SelectContext
         0     0% 93.46%  9728.11kB  0.63%  github.com/jmoiron/sqlx.SelectContext
         0     0% 93.46%  8244.74kB  0.53%  github.com/pressly/goose/v3.(*Migration).UpContext (inline)
         0     0% 93.46%  8438.73kB  0.54%  github.com/pressly/goose/v3.UpContext (inline)
         0     0% 93.46%  7904.66kB  0.51%  go.uber.org/zap.(*Logger).With
         0     0% 93.46%  7927.74kB  0.51%  go.uber.org/zap/buffer.Pool.Get
         0     0% 93.46% 10890.35kB   0.7%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0% 93.46%  7834.34kB  0.51%  go.uber.org/zap/zapcore.(*jsonEncoder).clone
         0     0% 93.46%  7816.34kB   0.5%  io.Copy (inline)
         0     0% 93.46% 10025.02kB  0.65%  main.main
         0     0% 93.46% 10025.02kB  0.65%  main.run
         0     0% 93.46%  8077.70kB  0.52%  net/http.(*chunkWriter).Write
         0     0% 93.46%  8664.33kB  0.56%  net/http.(*response).finishRequest
         0     0% 93.46% 30766.26kB  1.98%  net/http.HandlerFunc.ServeHTTP
         0     0% 93.46% 1497664.87kB 96.56%  net/http.serverHandler.ServeHTTP
         0     0% 93.46% 10025.02kB  0.65%  runtime.main
         0     0% 93.46% 34156.65kB  2.20%  sync.(*Pool).Get
```
