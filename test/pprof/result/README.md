# Result

Collection: `go run ./cmd/pproftest/ -d test/pprof/result -t 300 -m 1`

## Agent

### CPU

Top: `go tool pprof -top agent-cpu.pprof`

```
File: agent-3561658577
Type: cpu
Time: 2026-04-16 19:32:25 MSK
Duration: 299.04s, Total samples = 2.88s ( 0.96%)
Showing nodes accounting for 2.68s, 93.06% of 2.88s total
Dropped 112 nodes (cum <= 0.01s)
      flat  flat%   sum%        cum   cum%
     0.40s 13.89% 13.89%      0.40s 13.89%  syscall.syscall
     0.37s 12.85% 26.74%      0.37s 12.85%  runtime.pthread_cond_signal
     0.31s 10.76% 37.50%      0.31s 10.76%  runtime.usleep
     0.29s 10.07% 47.57%      0.29s 10.07%  runtime.pthread_cond_wait
     0.22s  7.64% 55.21%      0.22s  7.64%  runtime.kevent
     0.18s  6.25% 61.46%      0.18s  6.25%  runtime.cgocall
     0.13s  4.51% 65.97%      0.28s  9.72%  runtime.pcvalue
     0.11s  3.82% 69.79%      0.11s  3.82%  runtime.(*mspan).specialFindSplicePoint (inline)
     0.09s  3.12% 72.92%      0.09s  3.12%  runtime.memclrNoHeapPointers
     0.07s  2.43% 75.35%      0.12s  4.17%  runtime.step
     0.05s  1.74% 77.08%      0.05s  1.74%  runtime.findfunc
     0.05s  1.74% 78.82%      0.05s  1.74%  runtime.madvise
     0.05s  1.74% 80.56%      0.05s  1.74%  runtime.readvarint (inline)
     0.04s  1.39% 81.94%      0.15s  5.21%  runtime.(*unwinder).resolveInternal
     0.04s  1.39% 83.33%      0.04s  1.39%  runtime.pthread_kill
     0.03s  1.04% 84.38%      0.03s  1.04%  runtime.memmove
     0.03s  1.04% 85.42%      0.41s 14.24%  runtime.tracebackPCs
     0.02s  0.69% 86.11%      0.02s  0.69%  internal/runtime/atomic.(*UnsafePointer).Load (inline)
     0.02s  0.69% 86.81%      0.02s  0.69%  runtime.(*moduledata).textAddr
     0.02s  0.69% 87.50%      0.18s  6.25%  runtime.(*unwinder).next
     0.02s  0.69% 88.19%      0.03s  1.04%  runtime.(*unwinder).symPC
     0.02s  0.69% 88.89%      0.02s  0.69%  runtime.libcCall
     0.02s  0.69% 89.58%      0.35s 12.15%  runtime.setprofilebucket
     0.01s  0.35% 89.93%      0.02s  0.69%  github.com/ebitengine/purego.RegisterFunc
     0.01s  0.35% 90.28%      0.02s  0.69%  net/http.(*Transport).getConn
     0.01s  0.35% 90.62%      0.03s  1.04%  net/http.ReadResponse
     0.01s  0.35% 90.97%      0.02s  0.69%  runtime.(*inlineUnwinder).next
     0.01s  0.35% 91.32%      0.10s  3.47%  runtime.goschedImpl
     0.01s  0.35% 91.67%      0.15s  5.21%  runtime.newInlineUnwinder
     0.01s  0.35% 92.01%      0.38s 13.19%  runtime.semawakeup
     0.01s  0.35% 92.36%      0.04s  1.39%  runtime.stkbucket
     0.01s  0.35% 92.71%      0.02s  0.69%  strconv.genericFtoa
     0.01s  0.35% 93.06%      0.02s  0.69%  sync.(*Pool).Get
         0     0% 93.06%      0.39s 13.54%  bufio.(*Writer).Flush
         0     0% 93.06%      0.04s  1.39%  compress/flate.(*Writer).Close (inline)
         0     0% 93.06%      0.04s  1.39%  compress/flate.(*compressor).close
         0     0% 93.06%      0.03s  1.04%  compress/flate.(*compressor).encSpeed
         0     0% 93.06%      0.02s  0.69%  compress/gzip.(*Reader).Reset
         0     0% 93.06%      0.04s  1.39%  compress/gzip.(*Writer).Close
         0     0% 93.06%      0.02s  0.69%  compress/gzip.NewReader (inline)
         0     0% 93.06%      0.03s  1.04%  context.WithCancelCause
         0     0% 93.06%      0.02s  0.69%  context.WithDeadline (inline)
         0     0% 93.06%      0.02s  0.69%  context.WithDeadlineCause
         0     0% 93.06%      0.02s  0.69%  context.WithTimeout
         0     0% 93.06%      0.02s  0.69%  crypto/hmac.New
         0     0% 93.06%      0.02s  0.69%  crypto/hmac.New.UnwrapNew[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }].func1
         0     0% 93.06%      0.02s  0.69%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
         0     0% 93.06%      0.02s  0.69%  crypto/internal/fips140/sha256.New (inline)
         0     0% 93.06%      0.02s  0.69%  crypto/sha256.New
         0     0% 93.06%      0.04s  1.39%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*agent).Run.(*agent).launchReporter.func2
         0     0% 93.06%      0.04s  1.39%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*agent).doReport
         0     0% 93.06%      0.27s  9.38%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*collector).Collect.func1
         0     0% 93.06%      0.27s  9.38%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*collector).collectFromSource
         0     0% 93.06%      0.03s  1.04%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).getSendableMetrics
         0     0% 93.06%      0.04s  1.39%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).metricBatcher.func1
         0     0% 93.06%      0.80s 27.78%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).processBatches.func1
         0     0% 93.06%      0.80s 27.78%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).reportBatch
         0     0% 93.06%      0.80s 27.78%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).reportWorker (inline)
         0     0% 93.06%      0.77s 26.74%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).SendBatch
         0     0% 93.06%      0.10s  3.47%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).compressBody
         0     0% 93.06%      0.15s  5.21%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).prepareBody
         0     0% 93.06%      0.41s 14.24%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendSingleRequest
         0     0% 93.06%      0.66s 22.92%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendWithRetries
         0     0% 93.06%      0.41s 14.24%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendWithRetries.func1
         0     0% 93.06%      0.05s  1.74%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).signBody
         0     0% 93.06%      0.05s  1.74%  github.com/bq2cd/yp-go-metrics/internal/app/agent.runPeriodicTask
         0     0% 93.06%      0.02s  0.69%  github.com/bq2cd/yp-go-metrics/internal/app/agent/source/memstats.(*source).ReadMetrics
         0     0% 93.06%      0.25s  8.68%  github.com/bq2cd/yp-go-metrics/internal/app/agent/source/psutil.(*source).ReadMetrics
         0     0% 93.06%      0.08s  2.78%  github.com/bq2cd/yp-go-metrics/internal/app/agent/source/psutil.(*source).readCPUMetrics
         0     0% 93.06%      0.17s  5.90%  github.com/bq2cd/yp-go-metrics/internal/app/agent/source/psutil.(*source).readMemoryMetrics
         0     0% 93.06%      0.04s  1.39%  github.com/bq2cd/yp-go-metrics/internal/model.MetricSet.Upsert (inline)
         0     0% 93.06%      0.04s  1.39%  github.com/bq2cd/yp-go-metrics/internal/model.NewMetricSet (inline)
         0     0% 93.06%      0.04s  1.39%  github.com/bq2cd/yp-go-metrics/internal/model.NewMetricSetWithStrategy
         0     0% 93.06%      0.02s  0.69%  github.com/bq2cd/yp-go-metrics/pkg/bufpool.(*Pool).Get
         0     0% 93.06%      0.04s  1.39%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.(*Writer).Close
         0     0% 93.06%      0.04s  1.39%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.(*Writer).Close.func1
         0     0% 93.06%      0.02s  0.69%  github.com/bq2cd/yp-go-metrics/pkg/hmacsigner.(*hmacSigner).Sign
         0     0% 93.06%      0.06s  2.08%  github.com/bq2cd/yp-go-metrics/pkg/log.(*baseLogger).With
         0     0% 93.06%      0.05s  1.74%  github.com/bq2cd/yp-go-metrics/pkg/log.(*zapLogger).with
         0     0% 93.06%      0.02s  0.69%  github.com/bq2cd/yp-go-metrics/pkg/log.Str (inline)
         0     0% 93.06%      0.05s  1.74%  github.com/bq2cd/yp-go-metrics/pkg/periodictask.(*timerTask).Run
         0     0% 93.06%      0.45s 15.62%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.*uint8]).Do
         0     0% 93.06%      0.06s  2.08%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.NewRetrier[go.shape.*uint8]
         0     0% 93.06%      0.02s  0.69%  github.com/ebitengine/purego.Dlopen
         0     0% 93.06%      0.02s  0.69%  github.com/ebitengine/purego.Dlsym
         0     0% 93.06%      0.20s  6.94%  github.com/ebitengine/purego.RegisterFunc.func4
         0     0% 93.06%      0.04s  1.39%  github.com/ebitengine/purego.RegisterLibFunc
         0     0% 93.06%      0.02s  0.69%  github.com/ebitengine/purego.loadSymbol (inline)
         0     0% 93.06%      0.03s  1.04%  github.com/go-resty/resty/v2.(*Client).R (inline)
         0     0% 93.06%      0.32s 11.11%  github.com/go-resty/resty/v2.(*Client).execute
         0     0% 93.06%      0.12s  4.17%  github.com/go-resty/resty/v2.(*Client).executeBefore
         0     0% 93.06%      0.32s 11.11%  github.com/go-resty/resty/v2.(*Request).Execute
         0     0% 93.06%      0.32s 11.11%  github.com/go-resty/resty/v2.(*Request).Send (inline)
         0     0% 93.06%      0.06s  2.08%  github.com/go-resty/resty/v2.(*Request).SetHeader (inline)
         0     0% 93.06%      0.06s  2.08%  github.com/go-resty/resty/v2.(*Request).SetHeaders
         0     0% 93.06%      0.08s  2.78%  github.com/go-resty/resty/v2.createHTTPRequest
         0     0% 93.06%      0.04s  1.39%  github.com/go-resty/resty/v2.parseRequestURL
         0     0% 93.06%      0.02s  0.69%  github.com/go-resty/resty/v2.readAllWithLimit
         0     0% 93.06%      0.06s  2.08%  github.com/goccy/go-json.(*Encoder).Encode (inline)
         0     0% 93.06%      0.06s  2.08%  github.com/goccy/go-json.(*Encoder).EncodeWithOption
         0     0% 93.06%      0.06s  2.08%  github.com/goccy/go-json.(*Encoder).encodeWithOption
         0     0% 93.06%      0.02s  0.69%  github.com/goccy/go-json.Unmarshal (inline)
         0     0% 93.06%      0.06s  2.08%  github.com/goccy/go-json.encode
         0     0% 93.06%      0.05s  1.74%  github.com/goccy/go-json.encodeRunCode
         0     0% 93.06%      0.02s  0.69%  github.com/goccy/go-json.unmarshal
         0     0% 93.06%      0.02s  0.69%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).Decode
         0     0% 93.06%      0.02s  0.69%  github.com/goccy/go-json/internal/encoder.AppendFloat64
         0     0% 93.06%      0.05s  1.74%  github.com/goccy/go-json/internal/encoder/vm.Run
         0     0% 93.06%      0.08s  2.78%  github.com/shirou/gopsutil/v4/cpu.Percent (inline)
         0     0% 93.06%      0.08s  2.78%  github.com/shirou/gopsutil/v4/cpu.PercentWithContext
         0     0% 93.06%      0.08s  2.78%  github.com/shirou/gopsutil/v4/cpu.TimesWithContext
         0     0% 93.06%      0.08s  2.78%  github.com/shirou/gopsutil/v4/cpu.allCPUTimes
         0     0% 93.06%      0.08s  2.78%  github.com/shirou/gopsutil/v4/cpu.percentUsedFromLastCallWithContext
         0     0% 93.06%      0.03s  1.04%  github.com/shirou/gopsutil/v4/internal/common.GetFunc[go.shape.func int] (inline)
         0     0% 93.06%      0.03s  1.04%  github.com/shirou/gopsutil/v4/internal/common.NewLibrary
         0     0% 93.06%      0.17s  5.90%  github.com/shirou/gopsutil/v4/mem.VirtualMemory (inline)
         0     0% 93.06%      0.17s  5.90%  github.com/shirou/gopsutil/v4/mem.VirtualMemoryWithContext
         0     0% 93.06%      0.04s  1.39%  go.uber.org/zap.(*Logger).With
         0     0% 93.06%      0.02s  0.69%  go.uber.org/zap/buffer.Pool.Get
         0     0% 93.06%      0.02s  0.69%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0% 93.06%      0.03s  1.04%  go.uber.org/zap/zapcore.(*ioCore).With
         0     0% 93.06%      0.03s  1.04%  go.uber.org/zap/zapcore.(*ioCore).clone (inline)
         0     0% 93.06%      0.02s  0.69%  go.uber.org/zap/zapcore.(*jsonEncoder).Clone
         0     0% 93.06%      0.02s  0.69%  go.uber.org/zap/zapcore.(*jsonEncoder).clone
         0     0% 93.06%      0.04s  1.39%  go.uber.org/zap/zapcore.(*sampler).With
         0     0% 93.06%      1.12s 38.89%  golang.org/x/sync/errgroup.(*Group).Go.func1
         0     0% 93.06%      0.02s  0.69%  internal/bytealg.MakeNoZero
         0     0% 93.06%      0.39s 13.54%  internal/poll.(*FD).Write
         0     0% 93.06%      0.40s 13.89%  internal/poll.ignoringEINTRIO (inline)
         0     0% 93.06%      0.04s  1.39%  internal/runtime/maps.(*Map).growToSmall
         0     0% 93.06%      0.04s  1.39%  internal/runtime/maps.(*Map).growToTable
         0     0% 93.06%      0.03s  1.04%  internal/runtime/maps.(*table).reset
         0     0% 93.06%      0.05s  1.74%  internal/runtime/maps.NewEmptyMap (inline)
         0     0% 93.06%      0.03s  1.04%  internal/runtime/maps.NewMap
         0     0% 93.06%      0.07s  2.43%  internal/runtime/maps.newGroups (inline)
         0     0% 93.06%      0.04s  1.39%  internal/runtime/maps.newTable
         0     0% 93.06%      0.07s  2.43%  internal/runtime/maps.newarray
         0     0% 93.06%      0.05s  1.74%  io.Copy (inline)
         0     0% 93.06%      0.02s  0.69%  io.NopCloser (inline)
         0     0% 93.06%      0.02s  0.69%  io.ReadAll
         0     0% 93.06%      0.05s  1.74%  io.copyBuffer
         0     0% 93.06%      0.39s 13.54%  net.(*conn).Write
         0     0% 93.06%      0.39s 13.54%  net.(*netFD).Write
         0     0% 93.06%      0.18s  6.25%  net/http.(*Client).Do (inline)
         0     0% 93.06%      0.18s  6.25%  net/http.(*Client).do
         0     0% 93.06%      0.02s  0.69%  net/http.(*Client).makeHeadersCopier
         0     0% 93.06%      0.16s  5.56%  net/http.(*Client).send
         0     0% 93.06%      0.02s  0.69%  net/http.(*Request).write
         0     0% 93.06%      0.14s  4.86%  net/http.(*Transport).RoundTrip
         0     0% 93.06%      0.14s  4.86%  net/http.(*Transport).roundTrip
         0     0% 93.06%      0.02s  0.69%  net/http.(*cancelTimerBody).Read
         0     0% 93.06%      0.02s  0.69%  net/http.(*gzipReader).Read
         0     0% 93.06%      0.04s  1.39%  net/http.(*persistConn).readLoop
         0     0% 93.06%      0.03s  1.04%  net/http.(*persistConn).readResponse
         0     0% 93.06%      0.05s  1.74%  net/http.(*persistConn).roundTrip
         0     0% 93.06%      0.41s 14.24%  net/http.(*persistConn).writeLoop
         0     0% 93.06%      0.02s  0.69%  net/http.(*transportRequest).extraHeaders (inline)
         0     0% 93.06%      0.02s  0.69%  net/http.Header.Clone (inline)
         0     0% 93.06%      0.07s  2.43%  net/http.Header.Set (inline)
         0     0% 93.06%      0.07s  2.43%  net/http.NewRequest (inline)
         0     0% 93.06%      0.07s  2.43%  net/http.NewRequestWithContext
         0     0% 93.06%      0.02s  0.69%  net/http.cloneOrMakeHeader
         0     0% 93.06%      0.39s 13.54%  net/http.persistConnWriter.Write
         0     0% 93.06%      0.15s  5.21%  net/http.send
         0     0% 93.06%      0.02s  0.69%  net/http.setupRewindBody (inline)
         0     0% 93.06%      0.05s  1.74%  net/textproto.CanonicalMIMEHeaderKey
         0     0% 93.06%      0.07s  2.43%  net/textproto.MIMEHeader.Set (inline)
         0     0% 93.06%      0.05s  1.74%  net/textproto.canonicalMIMEHeaderKey
         0     0% 93.06%      0.02s  0.69%  net/url.(*URL).String
         0     0% 93.06%      0.04s  1.39%  net/url.Parse
         0     0% 93.06%      0.04s  1.39%  net/url.parse
         0     0% 93.06%      0.04s  1.39%  reflect.unsafe_New
         0     0% 93.06%      0.15s  5.21%  runtime.(*inlineUnwinder).resolveInternal (inline)
         0     0% 93.06%      0.03s  1.04%  runtime.(*mcache).nextFree
         0     0% 93.06%      0.03s  1.04%  runtime.(*mcache).prepareForSweep
         0     0% 93.06%      0.03s  1.04%  runtime.(*mcache).refill
         0     0% 93.06%      0.03s  1.04%  runtime.(*mcache).releaseAll
         0     0% 93.06%      0.02s  0.69%  runtime.(*mcentral).cacheSpan
         0     0% 93.06%      0.04s  1.39%  runtime.(*mcentral).uncacheSpan
         0     0% 93.06%      0.03s  1.04%  runtime.(*mheap).alloc.func1
         0     0% 93.06%      0.02s  0.69%  runtime.(*mheap).allocManual
         0     0% 93.06%      0.05s  1.74%  runtime.(*mheap).allocSpan
         0     0% 93.06%      0.02s  0.69%  runtime.(*stkframe).getStackMap
         0     0% 93.06%      0.05s  1.74%  runtime.(*sweepLocked).sweep
         0     0% 93.06%      0.04s  1.39%  runtime.(*unwinder).initAt
         0     0% 93.06%      0.02s  0.69%  runtime.acquirep
         0     0% 93.06%      0.11s  3.82%  runtime.addspecial
         0     0% 93.06%      0.02s  0.69%  runtime.adjustframe
         0     0% 93.06%      0.46s 15.97%  runtime.callers
         0     0% 93.06%      0.45s 15.62%  runtime.callers.func1
         0     0% 93.06%      0.02s  0.69%  runtime.concatstrings
         0     0% 93.06%      0.03s  1.04%  runtime.convT
         0     0% 93.06%      0.02s  0.69%  runtime.convTstring
         0     0% 93.06%      0.08s  2.78%  runtime.copystack
         0     0% 93.06%      0.02s  0.69%  runtime.entersyscall_sysmon
         0     0% 93.06%      0.47s 16.32%  runtime.findRunnable
         0     0% 93.06%      0.05s  1.74%  runtime.freeSpecial
         0     0% 93.06%      0.02s  0.69%  runtime.funcInfo.entry (inline)
         0     0% 93.06%      0.11s  3.82%  runtime.funcspdelta (inline)
         0     0% 93.06%      0.19s  6.60%  runtime.goexit0
         0     0% 93.06%      0.10s  3.47%  runtime.gopreempt_m (inline)
         0     0% 93.06%      0.02s  0.69%  runtime.growslice
         0     0% 93.06%      0.38s 13.19%  runtime.lock (inline)
         0     0% 93.06%      0.38s 13.19%  runtime.lock2
         0     0% 93.06%      0.38s 13.19%  runtime.lockWithRank (inline)
         0     0% 93.06%      0.25s  8.68%  runtime.mPark (inline)
         0     0% 93.06%      0.70s 24.31%  runtime.mProf_Malloc
         0     0% 93.06%      0.35s 12.15%  runtime.mProf_Malloc.func1
         0     0% 93.06%      0.04s  1.39%  runtime.makechan
         0     0% 93.06%      0.04s  1.39%  runtime.makemap
         0     0% 93.06%      0.05s  1.74%  runtime.makemap_small
         0     0% 93.06%      0.13s  4.51%  runtime.makeslice
         0     0% 93.06%      0.84s 29.17%  runtime.mallocgc
         0     0% 93.06%      0.05s  1.74%  runtime.mallocgcLarge
         0     0% 93.06%      0.11s  3.82%  runtime.mallocgcSmallNoscan
         0     0% 93.06%      0.07s  2.43%  runtime.mallocgcSmallScanHeader
         0     0% 93.06%      0.54s 18.75%  runtime.mallocgcSmallScanNoHeader
         0     0% 93.06%      0.07s  2.43%  runtime.mallocgcTiny
         0     0% 93.06%      0.04s  1.39%  runtime.mapassign
         0     0% 93.06%      0.04s  1.39%  runtime.mapassign_faststr
         0     0% 93.06%      0.55s 19.10%  runtime.mcall
         0     0% 93.06%      0.05s  1.74%  runtime.memclrNoHeapPointersChunked
         0     0% 93.06%      0.11s  3.82%  runtime.morestack
         0     0% 93.06%      0.22s  7.64%  runtime.netpoll
         0     0% 93.06%      0.07s  2.43%  runtime.newarray
         0     0% 93.06%      0.41s 14.24%  runtime.newobject
         0     0% 93.06%      0.18s  6.25%  runtime.newstack
         0     0% 93.06%      0.25s  8.68%  runtime.notesleep
         0     0% 93.06%      0.34s 11.81%  runtime.notewakeup
         0     0% 93.06%      0.31s 10.76%  runtime.osyield (inline)
         0     0% 93.06%      0.36s 12.50%  runtime.park_m
         0     0% 93.06%      0.02s  0.69%  runtime.pcdatavalue
         0     0% 93.06%      0.15s  5.21%  runtime.pcdatavalue1
         0     0% 93.06%      0.04s  1.39%  runtime.preemptM
         0     0% 93.06%      0.04s  1.39%  runtime.preemptall
         0     0% 93.06%      0.04s  1.39%  runtime.preemptone
         0     0% 93.06%      0.70s 24.31%  runtime.profilealloc
         0     0% 93.06%      0.02s  0.69%  runtime.pthread_mutex_unlock
         0     0% 93.06%      0.02s  0.69%  runtime.rawstring (inline)
         0     0% 93.06%      0.02s  0.69%  runtime.rawstringtmp
         0     0% 93.06%      0.25s  8.68%  runtime.ready
         0     0% 93.06%      0.02s  0.69%  runtime.readyWithTime.goready.func1
         0     0% 93.06%      0.07s  2.43%  runtime.resetspinning
         0     0% 93.06%      0.55s 19.10%  runtime.schedule
         0     0% 93.06%      0.32s 11.11%  runtime.semasleep
         0     0% 93.06%      0.22s  7.64%  runtime.send.goready.func1
         0     0% 93.06%      0.04s  1.39%  runtime.signalM (inline)
         0     0% 93.06%      0.04s  1.39%  runtime.slicebytetostring
         0     0% 93.06%      0.02s  0.69%  runtime.stackalloc
         0     0% 93.06%      0.02s  0.69%  runtime.startTheWorld.func1
         0     0% 93.06%      0.03s  1.04%  runtime.startTheWorldWithSema
         0     0% 93.06%      0.31s 10.76%  runtime.startm
         0     0% 93.06%      0.05s  1.74%  runtime.stopTheWorld.func1
         0     0% 93.06%      0.05s  1.74%  runtime.stopTheWorldWithSema
         0     0% 93.06%      0.26s  9.03%  runtime.stopm
         0     0% 93.06%      0.05s  1.74%  runtime.sysUsed (inline)
         0     0% 93.06%      0.05s  1.74%  runtime.sysUsedOS (inline)
         0     0% 93.06%      1.19s 41.32%  runtime.systemstack
         0     0% 93.06%      0.04s  1.39%  runtime.unlock (inline)
         0     0% 93.06%      0.04s  1.39%  runtime.unlock2
         0     0% 93.06%      0.04s  1.39%  runtime.unlock2Wake
         0     0% 93.06%      0.04s  1.39%  runtime.unlockWithRank (inline)
         0     0% 93.06%      0.31s 10.76%  runtime.wakep
         0     0% 93.06%      0.02s  0.69%  strconv.AppendFloat (inline)
         0     0% 93.06%      0.02s  0.69%  strings.(*Builder).Grow
         0     0% 93.06%      0.02s  0.69%  strings.(*Builder).grow
         0     0% 93.06%      0.04s  1.39%  sync.(*Once).Do (inline)
         0     0% 93.06%      0.04s  1.39%  sync.(*Once).doSlow
         0     0% 93.06%      0.39s 13.54%  syscall.Write (inline)
         0     0% 93.06%      0.39s 13.54%  syscall.write
```

### Memory (alloc_space)

Top: `go tool pprof -top -sample_index=alloc_space agent-mem.pprof`

```
File: agent-3561658577
Type: alloc_space
Time: 2026-04-16 19:37:24 MSK
Showing nodes accounting for 140543.09kB, 92.47% of 151994.31kB total
Dropped 467 nodes (cum <= 759.97kB)
      flat  flat%   sum%        cum   cum%
33920.25kB 22.32% 22.32% 72596.14kB 47.76%  io.copyBuffer
   33856kB 22.27% 44.59%    33856kB 22.27%  compress/flate.(*dictDecoder).init (inline)
   21384kB 14.07% 58.66% 38672.11kB 25.44%  compress/flate.NewWriter (inline)
10560.23kB  6.95% 65.61% 17288.11kB 11.37%  compress/flate.(*compressor).init
 7968.06kB  5.24% 70.85% 41824.11kB 27.52%  compress/flate.NewReader
 7045.88kB  4.64% 75.49%  7045.88kB  4.64%  github.com/bq2cd/yp-go-metrics/internal/model.MetricSet.Upsert (inline)
    6600kB  4.34% 79.83%     6600kB  4.34%  compress/flate.newDeflateFast (inline)
 4347.56kB  2.86% 82.69%  4347.56kB  2.86%  bufio.NewReaderSize (inline)
 2187.28kB  1.44% 84.13%  2187.28kB  1.44%  go.uber.org/zap/internal/bufferpool.init.NewPool.func1
 1850.17kB  1.22% 85.34%  1850.17kB  1.22%  github.com/bq2cd/yp-go-metrics/internal/model.NewMetricSetWithStrategy
    1596kB  1.05% 86.39%     1596kB  1.05%  github.com/bq2cd/yp-go-metrics/internal/app/agent/source/memstats.(*memStats).ReadStats
    1228kB  0.81% 87.20%     1228kB  0.81%  regexp.(*bitState).reset
    1024kB  0.67% 87.88%     1024kB  0.67%  runtime/pprof.StartCPUProfile
     896kB  0.59% 88.47%      896kB  0.59%  net/http.init.func15
  828.12kB  0.54% 89.01%   843.14kB  0.55%  net/textproto.MIMEHeader.Set (inline)
  810.75kB  0.53% 89.54% 47693.61kB 31.38%  io.ReadAll
  727.38kB  0.48% 90.02% 46882.72kB 30.85%  compress/gzip.NewReader (inline)
  628.52kB  0.41% 90.44% 136774.50kB 89.99%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).SendBatch
  523.69kB  0.34% 90.78%  2119.69kB  1.39%  github.com/bq2cd/yp-go-metrics/internal/app/agent/source/memstats.(*source).ReadMetrics
  447.42kB  0.29% 91.08%  1479.59kB  0.97%  github.com/goccy/go-json.unmarshal
  383.03kB  0.25% 91.33%   813.66kB  0.54%  net/http.(*persistConn).roundTrip
  298.12kB   0.2% 91.52%   960.81kB  0.63%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).signBody
  279.11kB  0.18% 91.71%  1254.26kB  0.83%  net/http.(*persistConn).readLoop
  215.38kB  0.14% 91.85%  3069.47kB  2.02%  github.com/bq2cd/yp-go-metrics/pkg/log.(*zapLogger).with
  148.92kB 0.098% 91.95%   909.56kB   0.6%  net/http.ReadResponse
  132.56kB 0.087% 92.03%  2588.96kB  1.70%  go.uber.org/zap/zapcore.(*sampler).With
  132.50kB 0.087% 92.12%  1802.13kB  1.19%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.NewRetrier[go.shape.*uint8]
   99.42kB 0.065% 92.19%  2456.40kB  1.62%  go.uber.org/zap/zapcore.(*ioCore).clone (inline)
   82.81kB 0.054% 92.24%  2051.09kB  1.35%  net/http.(*Transport).roundTrip
   82.70kB 0.054% 92.30% 53347.17kB 35.10%  github.com/go-resty/resty/v2.(*Client).execute
   66.30kB 0.044% 92.34%  1241.77kB  0.82%  net/http.(*Request).write
   49.69kB 0.033% 92.37% 74005.33kB 48.69%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).prepareBody
   49.69kB 0.033% 92.40% 56136.41kB 36.93%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.*uint8]).Do
   33.14kB 0.022% 92.43%  3102.61kB  2.04%  github.com/bq2cd/yp-go-metrics/pkg/log.(*baseLogger).With
   33.06kB 0.022% 92.45%  2415.40kB  1.59%  net/http.send
   11.16kB 0.0073% 92.46% 141595.84kB 93.16%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).reportWorker (inline)
    8.28kB 0.0054% 92.46%  2953.68kB  1.94%  net/http.(*Client).do
    3.66kB 0.0024% 92.46%  2403.15kB  1.58%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).metricBatcher.func1
    2.91kB 0.0019% 92.47%  2903.20kB  1.91%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*collector).collectFromSource
    0.66kB 0.00043% 92.47%  1242.42kB  0.82%  net/http.(*persistConn).writeLoop
    0.50kB 0.00033% 92.47%  1123.80kB  0.74%  github.com/bq2cd/yp-go-metrics/pkg/periodictask.(*timerTask).Run
    0.09kB 6.2e-05% 92.47% 41824.20kB 27.52%  compress/gzip.(*Reader).readHeader
    0.05kB 3.1e-05% 92.47%   927.64kB  0.61%  github.com/go-resty/resty/v2.createHTTPRequest
    0.03kB 2.1e-05% 92.47%  1040.90kB  0.68%  github.com/bq2cd/yp-go-metrics/internal/app/cli.App[go.shape.struct { UpstreamURL net/url.URL; PollInterval time.Duration; ReportInterval time.Duration; HMACSecretKey []uint8; SenderPoolSize uint }].Run
         0     0% 92.47%  4331.19kB  2.85%  bufio.NewReader (inline)
         0     0% 92.47% 46155.34kB 30.37%  compress/gzip.(*Reader).Reset
         0     0% 92.47% 38675.80kB 25.45%  compress/gzip.(*Writer).Write
         0     0% 92.47%   921.76kB  0.61%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*agent).Run.(*agent).launchReporter.func2
         0     0% 92.47%   920.94kB  0.61%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*agent).doReport
         0     0% 92.47%  2903.20kB  1.91%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*collector).Collect.func1
         0     0% 92.47%  2402.10kB  1.58%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).getSendableMetrics
         0     0% 92.47% 141595.84kB 93.16%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).processBatches.func1
         0     0% 92.47% 141584.69kB 93.15%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).reportBatch
         0     0% 92.47%  2408.09kB  1.58%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).storeReported
         0     0% 92.47% 72991.02kB 48.02%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).compressBody
         0     0% 92.47% 54571.25kB 35.90%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendSingleRequest
         0     0% 92.47% 131968.34kB 86.82%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendWithRetries
         0     0% 92.47% 54571.25kB 35.90%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendWithRetries.func1
         0     0% 92.47%  1123.80kB  0.74%  github.com/bq2cd/yp-go-metrics/internal/app/agent.runPeriodicTask
         0     0% 92.47%  1024.04kB  0.67%  github.com/bq2cd/yp-go-metrics/internal/app/cli.(*profiler).MaybeStartProfiling
         0     0% 92.47%     1024kB  0.67%  github.com/bq2cd/yp-go-metrics/internal/app/cli.(*profiler).maybeStartCPUProfiling
         0     0% 92.47%  1850.17kB  1.22%  github.com/bq2cd/yp-go-metrics/internal/model.NewMetricSet (inline)
         0     0% 92.47% 38675.80kB 25.45%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.(*Writer).Write
         0     0% 92.47%  2138.92kB  1.41%  github.com/go-resty/resty/v2.(*Client).executeBefore
         0     0% 92.47% 53347.17kB 35.10%  github.com/go-resty/resty/v2.(*Request).Execute
         0     0% 92.47% 53347.17kB 35.10%  github.com/go-resty/resty/v2.(*Request).Send (inline)
         0     0% 92.47%  1250.52kB  0.82%  github.com/go-resty/resty/v2.IsJSONType (inline)
         0     0% 92.47%   805.44kB  0.53%  github.com/go-resty/resty/v2.parseRequestHeader
         0     0% 92.47% 47693.61kB 31.38%  github.com/go-resty/resty/v2.readAllWithLimit
         0     0% 92.47%  1479.59kB  0.97%  github.com/goccy/go-json.Unmarshal (inline)
         0     0% 92.47%  2854.09kB  1.88%  go.uber.org/zap.(*Logger).With
         0     0% 92.47%  2207.13kB  1.45%  go.uber.org/zap/buffer.Pool.Get
         0     0% 92.47%  2187.28kB  1.44%  go.uber.org/zap/internal/bufferpool.init.NewPool.New[go.shape.*uint8].func2
         0     0% 92.47%  2363.37kB  1.55%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0% 92.47%  2456.40kB  1.62%  go.uber.org/zap/zapcore.(*ioCore).With
         0     0% 92.47%  2356.98kB  1.55%  go.uber.org/zap/zapcore.(*jsonEncoder).Clone
         0     0% 92.47%  2359.20kB  1.55%  go.uber.org/zap/zapcore.(*jsonEncoder).clone
         0     0% 92.47% 145746.12kB 95.89%  golang.org/x/sync/errgroup.(*Group).Go.func1
         0     0% 92.47% 72595.84kB 47.76%  io.Copy (inline)
         0     0% 92.47%  1040.90kB  0.68%  main.main
         0     0% 92.47%  1040.90kB  0.68%  main.run
         0     0% 92.47%  2953.68kB  1.94%  net/http.(*Client).Do (inline)
         0     0% 92.47%  2415.40kB  1.59%  net/http.(*Client).send
         0     0% 92.47%  2051.09kB  1.35%  net/http.(*Transport).RoundTrip
         0     0% 92.47% 46882.86kB 30.85%  net/http.(*cancelTimerBody).Read
         0     0% 92.47% 46882.86kB 30.85%  net/http.(*gzipReader).Read
         0     0% 92.47%   909.56kB   0.6%  net/http.(*persistConn).readResponse
         0     0% 92.47%   914.67kB   0.6%  net/http.(*transferWriter).doBodyCopy
         0     0% 92.47%   939.52kB  0.62%  net/http.(*transferWriter).writeBody
         0     0% 92.47%   843.14kB  0.55%  net/http.Header.Set (inline)
         0     0% 92.47%   914.38kB   0.6%  net/http.getCopyBuf (inline)
         0     0% 92.47%  1254.52kB  0.83%  regexp.(*Regexp).MatchString (inline)
         0     0% 92.47%  1254.52kB  0.83%  regexp.(*Regexp).backtrack
         0     0% 92.47%  1254.52kB  0.83%  regexp.(*Regexp).doExecute
         0     0% 92.47%  1254.52kB  0.83%  regexp.(*Regexp).doMatch (inline)
         0     0% 92.47%  1040.90kB  0.68%  runtime.main
         0     0% 92.47%   780.16kB  0.51%  sync.(*Once).Do (inline)
         0     0% 92.47%   780.16kB  0.51%  sync.(*Once).doSlow
         0     0% 92.47%  3530.95kB  2.32%  sync.(*Pool).Get
```

## Server

### CPU

Top: `go tool pprof -top server-cpu.pprof`

```
File: server-222229620
Type: cpu
Time: 2026-04-16 19:32:24 MSK
Duration: 299.36s, Total samples = 7.45s ( 2.49%)
Showing nodes accounting for 7.06s, 94.77% of 7.45s total
Dropped 276 nodes (cum <= 0.04s)
      flat  flat%   sum%        cum   cum%
     2.18s 29.26% 29.26%      2.18s 29.26%  runtime.pthread_cond_signal
     1.03s 13.83% 43.09%      1.03s 13.83%  syscall.syscall
     0.99s 13.29% 56.38%      0.99s 13.29%  runtime.usleep
     0.78s 10.47% 66.85%      0.78s 10.47%  runtime.kevent
     0.72s  9.66% 76.51%      0.72s  9.66%  runtime.pthread_cond_wait
     0.32s  4.30% 80.81%      0.32s  4.30%  runtime.(*mspan).specialFindSplicePoint (inline)
     0.20s  2.68% 83.49%      0.20s  2.68%  runtime.memclrNoHeapPointers
     0.12s  1.61% 85.10%      0.20s  2.68%  runtime.step
     0.08s  1.07% 86.17%      0.08s  1.07%  runtime.madvise
     0.08s  1.07% 87.25%      0.42s  5.64%  runtime.pcvalue
     0.08s  1.07% 88.32%      0.08s  1.07%  runtime.readvarint (inline)
     0.07s  0.94% 89.26%      0.07s  0.94%  runtime.acquirem (inline)
     0.07s  0.94% 90.20%      0.07s  0.94%  runtime.findfunc
     0.06s  0.81% 91.01%      0.06s  0.81%  runtime.(*moduledata).textAddr
     0.05s  0.67% 91.68%      1.18s 15.84%  runtime.setprofilebucket
     0.04s  0.54% 92.21%      0.04s  0.54%  syscall.rawSyscall
     0.03s   0.4% 92.62%      0.09s  1.21%  runtime.funcInfo.entry (inline)
     0.03s   0.4% 93.02%      0.94s 12.62%  runtime.lock2
     0.03s   0.4% 93.42%      0.62s  8.32%  runtime.tracebackPCs
     0.02s  0.27% 93.69%      0.11s  1.48%  runtime.unlock2
     0.01s  0.13% 93.83%      0.04s  0.54%  compress/gzip.(*Reader).Reset
     0.01s  0.13% 93.96%      0.04s  0.54%  io.ReadAll
     0.01s  0.13% 94.09%      0.12s  1.61%  net/http.(*conn).readRequest
     0.01s  0.13% 94.23%      0.26s  3.49%  runtime.(*unwinder).next
     0.01s  0.13% 94.36%      0.18s  2.42%  runtime.(*unwinder).resolveInternal
     0.01s  0.13% 94.50%      0.63s  8.46%  runtime.callers.func1
     0.01s  0.13% 94.63%      0.17s  2.28%  runtime.funcspdelta (inline)
     0.01s  0.13% 94.77%      0.26s  3.49%  runtime.newInlineUnwinder
         0     0% 94.77%      0.05s  0.67%  bufio.(*Reader).Peek
         0     0% 94.77%      0.05s  0.67%  bufio.(*Reader).fill
         0     0% 94.77%      0.07s  0.94%  bufio.(*Writer).Flush
         0     0% 94.77%      0.11s  1.48%  database/sql.(*DB).BeginTx
         0     0% 94.77%      0.11s  1.48%  database/sql.(*DB).BeginTx.func1
         0     0% 94.77%      0.51s  6.85%  database/sql.(*DB).QueryContext
         0     0% 94.77%      0.51s  6.85%  database/sql.(*DB).QueryContext.func1
         0     0% 94.77%      0.11s  1.48%  database/sql.(*DB).begin
         0     0% 94.77%      0.10s  1.34%  database/sql.(*DB).beginDC
         0     0% 94.77%      0.09s  1.21%  database/sql.(*DB).beginDC.func1
         0     0% 94.77%      0.21s  2.82%  database/sql.(*DB).conn
         0     0% 94.77%      0.08s  1.07%  database/sql.(*DB).execDC
         0     0% 94.77%      0.08s  1.07%  database/sql.(*DB).execDC.func2
         0     0% 94.77%      0.51s  6.85%  database/sql.(*DB).query
         0     0% 94.77%      0.31s  4.16%  database/sql.(*DB).queryDC
         0     0% 94.77%      0.31s  4.16%  database/sql.(*DB).queryDC.func1
         0     0% 94.77%      0.62s  8.32%  database/sql.(*DB).retry
         0     0% 94.77%      0.11s  1.48%  database/sql.(*Rows).Next
         0     0% 94.77%      0.10s  1.34%  database/sql.(*Rows).Next.func1
         0     0% 94.77%      0.10s  1.34%  database/sql.(*Rows).nextLocked
         0     0% 94.77%      0.08s  1.07%  database/sql.(*Tx).ExecContext
         0     0% 94.77%      0.09s  1.21%  database/sql.ctxDriverBegin
         0     0% 94.77%      0.08s  1.07%  database/sql.ctxDriverExec
         0     0% 94.77%      0.31s  4.16%  database/sql.ctxDriverQuery
         0     0% 94.77%      0.62s  8.32%  database/sql.withLock
         0     0% 94.77%      0.17s  2.28%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).createPeriodicTask.func2
         0     0% 94.77%      0.17s  2.28%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).dumpMetrics
         0     0% 94.77%      0.63s  8.46%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).launchBatchWriter.func1
         0     0% 94.77%      0.17s  2.28%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).launchMetricDumper.func1
         0     0% 94.77%      1.06s 14.23%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).ServeHTTP
         0     0% 94.77%      0.92s 12.35%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).retrieveMetrics
         0     0% 94.77%      0.07s  0.94%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).validateMetrics
         0     0% 94.77%      1.66s 22.28%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorMiddleware).Intercept
         0     0% 94.77%      0.04s  0.54%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorMiddleware).decompressRequest
         0     0% 94.77%      1.16s 15.57%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).Intercept
         0     0% 94.77%      0.06s  0.81%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).validateRequest
         0     0% 94.77%      0.04s  0.54%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).writeResponse
         0     0% 94.77%      1.60s 21.48%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*loggerMiddleware).Intercept
         0     0% 94.77%      1.71s 22.95%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*middlewareHandler).ServeHTTP
         0     0% 94.77%      1.06s 14.23%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*recovererMiddleware).Intercept
         0     0% 94.77%      1.71s 22.95%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*requestIDMiddleware).Intercept
         0     0% 94.77%      0.05s  0.67%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.generateRequestID
         0     0% 94.77%      0.05s  0.67%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.getOrGenerateRequestID
         0     0% 94.77%      1.71s 22.95%  github.com/bq2cd/yp-go-metrics/internal/handler/router.(*Router).ServeHTTP
         0     0% 94.77%      0.06s  0.81%  github.com/bq2cd/yp-go-metrics/internal/model.MetricKey.Compare (inline)
         0     0% 94.77%      0.04s  0.54%  github.com/bq2cd/yp-go-metrics/internal/model.NewAuditEvent
         0     0% 94.77%      0.04s  0.54%  github.com/bq2cd/yp-go-metrics/internal/model.NewAuditMetricName (inline)
         0     0% 94.77%      0.04s  0.54%  github.com/bq2cd/yp-go-metrics/internal/model.NewMetricSetWithStrategy
         0     0% 94.77%      0.10s  1.34%  github.com/bq2cd/yp-go-metrics/internal/repository/auditsink.(*fileSink).WriteEvent
         0     0% 94.77%      0.07s  0.94%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).GetAll
         0     0% 94.77%      1.04s 13.96%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).GetMulti
         0     0% 94.77%      0.46s  6.17%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).SetMulti
         0     0% 94.77%      1.05s 14.09%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByType
         0     0% 94.77%      1.10s 14.77%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByTypeWithRetries
         0     0% 94.77%      1.05s 14.09%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByTypeWithRetries.func1
         0     0% 94.77%      1.01s 13.56%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiForType
         0     0% 94.77%      0.44s  5.91%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMulti
         0     0% 94.77%      0.43s  5.77%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiForType
         0     0% 94.77%      0.46s  6.17%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiWithRetries
         0     0% 94.77%      0.44s  5.91%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiWithRetries.func1
         0     0% 94.77%      0.12s  1.61%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].ConvertMetrics
         0     0% 94.77%      0.05s  0.67%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].ConvertMetrics.Keys[go.shape.map[github.com/bq2cd/yp-go-metrics/internal/model.MetricKey]github.com/bq2cd/yp-go-metrics/internal/model.Metric,go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" },go.shape.struct { ID string "json:\"id\""; Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; Delta *int64 "json:\"delta,omitempty\""; Value *float64 "json:\"value,omitempty\""; Hash github.com/bq2cd/yp-go-metrics/internal/model.MetricHash "json:\"hash,omitempty\"" }].func2
         0     0% 94.77%      0.06s  0.81%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].ConvertMetrics.func1
         0     0% 94.77%      0.14s  1.88%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].Insert
         0     0% 94.77%      0.81s 10.87%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].Select
         0     0% 94.77%      0.20s  2.68%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value int64 "db:\"value\"" }].Select
         0     0% 94.77%      0.10s  1.34%  github.com/bq2cd/yp-go-metrics/internal/service.(*auditEventProcessor).processEvent.func1
         0     0% 94.77%      0.04s  0.54%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricAuditor).RecordMetricsUploaded
         0     0% 94.77%      0.06s  0.81%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricJSONEncoder).EncodeBatch
         0     0% 94.77%      0.13s  1.74%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricSnapshotter).DumpClose
         0     0% 94.77%      0.07s  0.94%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricStorer).RetrieveAll
         0     0% 94.77%      0.92s 12.35%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricStorer).RetrieveBatch
         0     0% 94.77%      0.63s  8.46%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).StartProcessing
         0     0% 94.77%      0.17s  2.28%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).accumulateCounters
         0     0% 94.77%      0.63s  8.46%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).processBatchTx
         0     0% 94.77%      0.04s  0.54%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.(*ReaderPool).Get
         0     0% 94.77%      0.14s  1.88%  github.com/bq2cd/yp-go-metrics/pkg/log.(*baseLogger).With
         0     0% 94.77%      0.32s  4.30%  github.com/bq2cd/yp-go-metrics/pkg/log.(*eventBuilder).Msg
         0     0% 94.77%      0.32s  4.30%  github.com/bq2cd/yp-go-metrics/pkg/log.(*zapLogger).log
         0     0% 94.77%      0.13s  1.74%  github.com/bq2cd/yp-go-metrics/pkg/log.(*zapLogger).with
         0     0% 94.77%      0.17s  2.28%  github.com/bq2cd/yp-go-metrics/pkg/periodictask.(*timerTask).Run
         0     0% 94.77%      1.08s 14.50%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.[]github.com/bq2cd/yp-go-metrics/internal/model.Metric]).Do
         0     0% 94.77%      0.45s  6.04%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.interface {}]).Do
         0     0% 94.77%      1.71s 22.95%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 94.77%      1.06s 14.23%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 94.77%      0.04s  0.54%  github.com/goccy/go-json.(*Decoder).Decode (inline)
         0     0% 94.77%      0.04s  0.54%  github.com/goccy/go-json.(*Decoder).DecodeWithOption
         0     0% 94.77%      0.06s  0.81%  github.com/goccy/go-json.(*Encoder).Encode (inline)
         0     0% 94.77%      0.06s  0.81%  github.com/goccy/go-json.(*Encoder).EncodeWithOption
         0     0% 94.77%      0.06s  0.81%  github.com/goccy/go-json.(*Encoder).encodeWithOption
         0     0% 94.77%      0.04s  0.54%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).DecodeStream
         0     0% 94.77%      0.16s  2.15%  github.com/huandu/go-sqlbuilder.(*Args).CompileWithFlavor
         0     0% 94.77%      0.13s  1.74%  github.com/huandu/go-sqlbuilder.(*Args).compileDigits
         0     0% 94.77%      0.13s  1.74%  github.com/huandu/go-sqlbuilder.(*Args).compileSuccessive
         0     0% 94.77%      0.04s  0.54%  github.com/huandu/go-sqlbuilder.(*Cond).In.func1
         0     0% 94.77%      0.06s  0.81%  github.com/huandu/go-sqlbuilder.(*InsertBuilder).BuildWithFlavor
         0     0% 94.77%      0.19s  2.55%  github.com/huandu/go-sqlbuilder.(*SelectBuilder).BuildWithFlavor
         0     0% 94.77%      0.11s  1.48%  github.com/huandu/go-sqlbuilder.(*WhereClause).BuildWithFlavor
         0     0% 94.77%      0.13s  1.74%  github.com/huandu/go-sqlbuilder.(*argsCompileContext).WriteValue
         0     0% 94.77%      0.09s  1.21%  github.com/huandu/go-sqlbuilder.(*clause).Build
         0     0% 94.77%      0.10s  1.34%  github.com/huandu/go-sqlbuilder.(*stringBuilder).WriteString (inline)
         0     0% 94.77%      0.04s  0.54%  github.com/huandu/go-sqlbuilder.(*stringBuilder).WriteStrings
         0     0% 94.77%      0.04s  0.54%  github.com/huandu/go-sqlbuilder.Flavor.NewSelectBuilder (inline)
         0     0% 94.77%      0.04s  0.54%  github.com/huandu/go-sqlbuilder.NewSelectBuilder
         0     0% 94.77%      0.04s  0.54%  github.com/huandu/go-sqlbuilder.newSelectBuilder
         0     0% 94.77%      0.05s  0.67%  github.com/huandu/go-sqlbuilder.newStringBuilder (inline)
         0     0% 94.77%      0.09s  1.21%  github.com/jackc/pgx/v5.(*Conn).BeginTx
         0     0% 94.77%      0.20s  2.68%  github.com/jackc/pgx/v5.(*Conn).Exec
         0     0% 94.77%      0.31s  4.16%  github.com/jackc/pgx/v5.(*Conn).Query
         0     0% 94.77%      0.20s  2.68%  github.com/jackc/pgx/v5.(*Conn).exec
         0     0% 94.77%      0.08s  1.07%  github.com/jackc/pgx/v5.(*Conn).execPrepared
         0     0% 94.77%      0.12s  1.61%  github.com/jackc/pgx/v5.(*Conn).execSimpleProtocol
         0     0% 94.77%      0.13s  1.74%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 94.77%      0.06s  0.81%  github.com/jackc/pgx/v5.ParseConfig (inline)
         0     0% 94.77%      0.06s  0.81%  github.com/jackc/pgx/v5.ParseConfigWithOptions
         0     0% 94.77%      0.13s  1.74%  github.com/jackc/pgx/v5.connect
         0     0% 94.77%      0.11s  1.48%  github.com/jackc/pgx/v5/pgconn.(*PgConn).Exec
         0     0% 94.77%      0.37s  4.97%  github.com/jackc/pgx/v5/pgconn.(*PgConn).ExecPrepared
         0     0% 94.77%      0.16s  2.15%  github.com/jackc/pgx/v5/pgconn.(*PgConn).enterPotentialWriteReadDeadlock (inline)
         0     0% 94.77%      0.04s  0.54%  github.com/jackc/pgx/v5/pgconn.(*PgConn).execExtendedPrefix
         0     0% 94.77%      0.33s  4.43%  github.com/jackc/pgx/v5/pgconn.(*PgConn).execExtendedSuffix
         0     0% 94.77%      0.40s  5.37%  github.com/jackc/pgx/v5/pgconn.(*PgConn).flushWithPotentialWriteReadDeadlock
         0     0% 94.77%      0.05s  0.67%  github.com/jackc/pgx/v5/pgconn.(*PgConn).peekMessage
         0     0% 94.77%      0.04s  0.54%  github.com/jackc/pgx/v5/pgconn.(*ResultReader).readUntilRowDescription
         0     0% 94.77%      0.13s  1.74%  github.com/jackc/pgx/v5/pgconn.ConnectConfig
         0     0% 94.77%      0.06s  0.81%  github.com/jackc/pgx/v5/pgconn.ParseConfigWithOptions
         0     0% 94.77%      0.13s  1.74%  github.com/jackc/pgx/v5/pgconn.connectOne
         0     0% 94.77%      0.13s  1.74%  github.com/jackc/pgx/v5/pgconn.connectPreferred
         0     0% 94.77%      0.04s  0.54%  github.com/jackc/pgx/v5/pgconn.defaultSettings
         0     0% 94.77%      0.07s  0.94%  github.com/jackc/pgx/v5/pgconn/ctxwatch.(*ContextWatcher).Watch
         0     0% 94.77%      0.05s  0.67%  github.com/jackc/pgx/v5/pgconn/internal/bgreader.(*BGReader).Read
         0     0% 94.77%      0.24s  3.22%  github.com/jackc/pgx/v5/pgproto3.(*Frontend).Flush
         0     0% 94.77%      0.05s  0.67%  github.com/jackc/pgx/v5/pgproto3.(*Frontend).Receive
         0     0% 94.77%      0.05s  0.67%  github.com/jackc/pgx/v5/pgproto3.(*chunkReader).Next
         0     0% 94.77%      0.09s  1.21%  github.com/jackc/pgx/v5/stdlib.(*Conn).BeginTx
         0     0% 94.77%      0.08s  1.07%  github.com/jackc/pgx/v5/stdlib.(*Conn).ExecContext
         0     0% 94.77%      0.31s  4.16%  github.com/jackc/pgx/v5/stdlib.(*Conn).QueryContext
         0     0% 94.77%      0.10s  1.34%  github.com/jackc/pgx/v5/stdlib.(*Rows).Next
         0     0% 94.77%      0.07s  0.94%  github.com/jackc/pgx/v5/stdlib.(*Rows).Next.func14
         0     0% 94.77%      0.19s  2.55%  github.com/jackc/pgx/v5/stdlib.(*driverConnector).Connect
         0     0% 94.77%      0.11s  1.48%  github.com/jmoiron/sqlx.(*DB).BeginTxx
         0     0% 94.77%      0.51s  6.85%  github.com/jmoiron/sqlx.(*DB).QueryxContext
         0     0% 94.77%      0.71s  9.53%  github.com/jmoiron/sqlx.(*DB).SelectContext
         0     0% 94.77%      0.71s  9.53%  github.com/jmoiron/sqlx.SelectContext
         0     0% 94.77%      0.20s  2.68%  github.com/jmoiron/sqlx.scanAll
         0     0% 94.77%      0.32s  4.30%  go.uber.org/zap.(*Logger).Log
         0     0% 94.77%      0.08s  1.07%  go.uber.org/zap.(*Logger).With
         0     0% 94.77%      0.04s  0.54%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0% 94.77%      0.30s  4.03%  go.uber.org/zap/zapcore.(*CheckedEntry).Write
         0     0% 94.77%      0.07s  0.94%  go.uber.org/zap/zapcore.(*ioCore).With
         0     0% 94.77%      0.30s  4.03%  go.uber.org/zap/zapcore.(*ioCore).Write
         0     0% 94.77%      0.06s  0.81%  go.uber.org/zap/zapcore.(*ioCore).clone (inline)
         0     0% 94.77%      0.04s  0.54%  go.uber.org/zap/zapcore.(*jsonEncoder).Clone
         0     0% 94.77%      0.04s  0.54%  go.uber.org/zap/zapcore.(*jsonEncoder).clone
         0     0% 94.77%      0.30s  4.03%  go.uber.org/zap/zapcore.(*lockedWriteSyncer).Write
         0     0% 94.77%      0.07s  0.94%  go.uber.org/zap/zapcore.(*sampler).With
         0     0% 94.77%      0.10s  1.34%  golang.org/x/sync/errgroup.(*Group).Go.func1
         0     0% 94.77%      0.15s  2.01%  internal/poll.(*FD).Read
         0     0% 94.77%      0.75s 10.07%  internal/poll.(*FD).Write
         0     0% 94.77%      0.90s 12.08%  internal/poll.ignoringEINTRIO (inline)
         0     0% 94.77%      0.07s  0.94%  internal/runtime/maps.NewMap
         0     0% 94.77%      0.04s  0.54%  internal/runtime/maps.newGroups (inline)
         0     0% 94.77%      0.04s  0.54%  internal/runtime/maps.newarray
         0     0% 94.77%      0.09s  1.21%  io.ReadAtLeast
         0     0% 94.77%      0.04s  0.54%  io.ReadFull (inline)
         0     0% 94.77%      0.07s  0.94%  net.(*Dialer).DialContext
         0     0% 94.77%      0.15s  2.01%  net.(*conn).Read
         0     0% 94.77%      0.32s  4.30%  net.(*conn).Write
         0     0% 94.77%      0.15s  2.01%  net.(*netFD).Read
         0     0% 94.77%      0.32s  4.30%  net.(*netFD).Write
         0     0% 94.77%      0.07s  0.94%  net.(*sysDialer).dialParallel
         0     0% 94.77%      0.07s  0.94%  net.(*sysDialer).dialSerial
         0     0% 94.77%      0.07s  0.94%  net.(*sysDialer).dialSingle
         0     0% 94.77%      0.07s  0.94%  net.(*sysDialer).dialTCP
         0     0% 94.77%      0.07s  0.94%  net.(*sysDialer).doDialTCP (inline)
         0     0% 94.77%      0.07s  0.94%  net.(*sysDialer).doDialTCPProto
         0     0% 94.77%      0.05s  0.67%  net.internetSocket
         0     0% 94.77%      0.05s  0.67%  net.socket
         0     0% 94.77%      0.04s  0.54%  net.sysSocket
         0     0% 94.77%      1.95s 26.17%  net/http.(*conn).serve
         0     0% 94.77%      0.05s  0.67%  net/http.(*connReader).Read
         0     0% 94.77%      0.04s  0.54%  net/http.(*connReader).backgroundRead
         0     0% 94.77%      0.07s  0.94%  net/http.(*response).finishRequest
         0     0% 94.77%      1.06s 14.23%  net/http.HandlerFunc.ServeHTTP
         0     0% 94.77%      0.07s  0.94%  net/http.checkConnErrorWriter.Write
         0     0% 94.77%      0.08s  1.07%  net/http.readRequest
         0     0% 94.77%      1.71s 22.95%  net/http.serverHandler.ServeHTTP
         0     0% 94.77%      0.07s  0.94%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0% 94.77%      0.07s  0.94%  net/textproto.readMIMEHeader
         0     0% 94.77%      0.43s  5.77%  os.(*File).Write
         0     0% 94.77%      0.43s  5.77%  os.(*File).write (inline)
         0     0% 94.77%      0.05s  0.67%  os.OpenFile
         0     0% 94.77%      0.04s  0.54%  os.Stat
         0     0% 94.77%      0.09s  1.21%  os.ignoringEINTR (inline)
         0     0% 94.77%      0.05s  0.67%  os.open (inline)
         0     0% 94.77%      0.05s  0.67%  os.openFileNolog
         0     0% 94.77%      0.05s  0.67%  os.openFileNolog.func1 (inline)
         0     0% 94.77%      0.04s  0.54%  os.statNolog
         0     0% 94.77%      0.04s  0.54%  os.statNolog.func1 (inline)
         0     0% 94.77%      0.04s  0.54%  reflect.Append
         0     0% 94.77%      0.04s  0.54%  reflect.Value.extendSlice
         0     0% 94.77%      0.06s  0.81%  reflect.unsafe_New
         0     0% 94.77%      0.25s  3.36%  runtime.(*inlineUnwinder).resolveInternal (inline)
         0     0% 94.77%      0.07s  0.94%  runtime.(*mcache).nextFree
         0     0% 94.77%      0.07s  0.94%  runtime.(*mcache).refill
         0     0% 94.77%      0.07s  0.94%  runtime.(*mcentral).cacheSpan
         0     0% 94.77%      0.06s  0.81%  runtime.(*mcentral).grow
         0     0% 94.77%      0.10s  1.34%  runtime.(*mheap).alloc.func1
         0     0% 94.77%      0.10s  1.34%  runtime.(*mheap).allocSpan
         0     0% 94.77%      0.04s  0.54%  runtime.(*mspan).initHeapBits
         0     0% 94.77%      0.08s  1.07%  runtime.(*sweepLocked).sweep
         0     0% 94.77%      0.14s  1.88%  runtime.(*timer).maybeAdd
         0     0% 94.77%      0.16s  2.15%  runtime.(*timer).modify
         0     0% 94.77%      0.16s  2.15%  runtime.(*timer).reset (inline)
         0     0% 94.77%      0.35s  4.70%  runtime.addspecial
         0     0% 94.77%      0.05s  0.67%  runtime.bgsweep
         0     0% 94.77%      0.63s  8.46%  runtime.callers
         0     0% 94.77%      0.08s  1.07%  runtime.convTstring
         0     0% 94.77%      1.46s 19.60%  runtime.findRunnable
         0     0% 94.77%      0.08s  1.07%  runtime.freeSpecial
         0     0% 94.77%      0.04s  0.54%  runtime.gcBgMarkWorker
         0     0% 94.77%      0.04s  0.54%  runtime.gcBgMarkWorker.func2
         0     0% 94.77%      0.04s  0.54%  runtime.gcDrain
         0     0% 94.77%      0.34s  4.56%  runtime.goexit0
         0     0% 94.77%      0.29s  3.89%  runtime.growslice
         0     0% 94.77%      0.94s 12.62%  runtime.lock (inline)
         0     0% 94.77%      0.94s 12.62%  runtime.lockWithRank (inline)
         0     0% 94.77%      0.66s  8.86%  runtime.mPark (inline)
         0     0% 94.77%      1.27s 17.05%  runtime.mProf_Malloc
         0     0% 94.77%      1.18s 15.84%  runtime.mProf_Malloc.func1
         0     0% 94.77%      0.07s  0.94%  runtime.makemap
         0     0% 94.77%      0.16s  2.15%  runtime.makeslice
         0     0% 94.77%      1.49s 20.00%  runtime.mallocgc
         0     0% 94.77%      0.25s  3.36%  runtime.mallocgcSmallNoscan
         0     0% 94.77%      0.05s  0.67%  runtime.mallocgcSmallScanHeader
         0     0% 94.77%      1.03s 13.83%  runtime.mallocgcSmallScanNoHeader
         0     0% 94.77%      0.14s  1.88%  runtime.mallocgcTiny
         0     0% 94.77%      0.04s  0.54%  runtime.markroot
         0     0% 94.77%      1.58s 21.21%  runtime.mcall
         0     0% 94.77%      0.62s  8.32%  runtime.netpoll
         0     0% 94.77%      0.16s  2.15%  runtime.netpollBreak (inline)
         0     0% 94.77%      0.04s  0.54%  runtime.newarray
         0     0% 94.77%      0.69s  9.26%  runtime.newobject
         0     0% 94.77%      1.39s 18.66%  runtime.newproc.func1
         0     0% 94.77%      0.66s  8.86%  runtime.notesleep
         0     0% 94.77%      2.11s 28.32%  runtime.notewakeup
         0     0% 94.77%      0.85s 11.41%  runtime.osyield (inline)
         0     0% 94.77%      1.24s 16.64%  runtime.park_m
         0     0% 94.77%      0.25s  3.36%  runtime.pcdatavalue1
         0     0% 94.77%      1.27s 17.05%  runtime.profilealloc
         0     0% 94.77%      0.61s  8.19%  runtime.ready
         0     0% 94.77%      0.12s  1.61%  runtime.resetspinning
         0     0% 94.77%      0.14s  1.88%  runtime.runqgrab
         0     0% 94.77%      0.14s  1.88%  runtime.runqsteal
         0     0% 94.77%      1.58s 21.21%  runtime.schedule
         0     0% 94.77%      0.72s  9.66%  runtime.semasleep
         0     0% 94.77%      2.18s 29.26%  runtime.semawakeup
         0     0% 94.77%      0.61s  8.19%  runtime.send.goready.func1
         0     0% 94.77%      0.09s  1.21%  runtime.slicebytetostring
         0     0% 94.77%      2.10s 28.19%  runtime.startm
         0     0% 94.77%      0.14s  1.88%  runtime.stealWork
         0     0% 94.77%      0.66s  8.86%  runtime.stopm
         0     0% 94.77%      0.05s  0.67%  runtime.sweepone
         0     0% 94.77%      0.08s  1.07%  runtime.sysUsed (inline)
         0     0% 94.77%      0.08s  1.07%  runtime.sysUsedOS (inline)
         0     0% 94.77%      3.98s 53.42%  runtime.systemstack
         0     0% 94.77%      0.11s  1.48%  runtime.unlock (inline)
         0     0% 94.77%      0.07s  0.94%  runtime.unlock2Wake
         0     0% 94.77%      0.11s  1.48%  runtime.unlockWithRank (inline)
         0     0% 94.77%      0.16s  2.15%  runtime.wakeNetPoller
         0     0% 94.77%      0.16s  2.15%  runtime.wakeNetpoll
         0     0% 94.77%      2.10s 28.19%  runtime.wakep
         0     0% 94.77%      0.06s  0.81%  slices.AppendSeq[go.shape.[]go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" },go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }] (inline)
         0     0% 94.77%      0.06s  0.81%  slices.Collect[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }] (inline)
         0     0% 94.77%      0.06s  0.81%  slices.SortStableFunc[go.shape.[]go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" },go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }] (inline)
         0     0% 94.77%      0.12s  1.61%  slices.SortedStableFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }]
         0     0% 94.77%      0.05s  0.67%  slices.SortedStableFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }].Collect[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }].AppendSeq[go.shape.[]go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" },go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }]-range1
         0     0% 94.77%      0.06s  0.81%  slices.insertionSortCmpFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }] (inline)
         0     0% 94.77%      0.06s  0.81%  slices.stableCmpFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }]
         0     0% 94.77%      0.11s  1.48%  strings.(*Builder).WriteString (inline)
         0     0% 94.77%      0.05s  0.67%  sync.(*Pool).Get
         0     0% 94.77%      0.05s  0.67%  syscall.Open
         0     0% 94.77%      0.15s  2.01%  syscall.Read (inline)
         0     0% 94.77%      0.04s  0.54%  syscall.Socket
         0     0% 94.77%      0.04s  0.54%  syscall.Stat
         0     0% 94.77%      0.75s 10.07%  syscall.Write (inline)
         0     0% 94.77%      0.15s  2.01%  syscall.read
         0     0% 94.77%      0.04s  0.54%  syscall.socket
         0     0% 94.77%      0.75s 10.07%  syscall.write
         0     0% 94.77%      0.16s  2.15%  time.(*Timer).Reset
         0     0% 94.77%      0.16s  2.15%  time.resetTimer
```

### Memory (alloc_space)

Top: `go tool pprof -top -sample_index=alloc_space server-mem.pprof`

```
File: server-222229620
Type: alloc_space
Time: 2026-04-16 19:37:24 MSK
Showing nodes accounting for 75220.05kB, 85.57% of 87901.54kB total
Dropped 840 nodes (cum <= 439.51kB)
      flat  flat%   sum%        cum   cum%
   11664kB 13.27% 13.27% 20709.75kB 23.56%  compress/flate.NewWriter (inline)
 8192.38kB  9.32% 22.59%  8192.38kB  9.32%  github.com/pressly/goose/v3/internal/sqlparser.init.func1
    5511kB  6.27% 28.86%     5511kB  6.27%  go.uber.org/zap/internal/bufferpool.init.NewPool.func1
    5440kB  6.19% 35.05%  9045.75kB 10.29%  compress/flate.(*compressor).init
 4734.31kB  5.39% 40.43%  4734.31kB  5.39%  github.com/bq2cd/yp-go-metrics/internal/model.MetricSet.GroupByType (inline)
 4003.69kB  4.55% 44.99%  4003.69kB  4.55%  bufio.NewReaderSize (inline)
    3400kB  3.87% 48.86%     3400kB  3.87%  compress/flate.newDeflateFast (inline)
 3049.51kB  3.47% 52.33%  3049.51kB  3.47%  github.com/bq2cd/yp-go-metrics/internal/model.NewMetricSetWithStrategy
 1268.66kB  1.44% 53.77%  1268.66kB  1.44%  strings.(*Builder).WriteString (inline)
 1251.69kB  1.42% 55.19%  1251.69kB  1.42%  github.com/bq2cd/yp-go-metrics/internal/model.MetricKey.Compare (inline)
    1152kB  1.31% 56.50%     1152kB  1.31%  runtime/pprof.StartCPUProfile
 1099.44kB  1.25% 57.75%  1882.72kB  2.14%  github.com/huandu/go-sqlbuilder.(*argsCompileContext).WriteValue
 1047.83kB  1.19% 58.95%  1047.83kB  1.19%  bytes.growSlice
  970.73kB  1.10% 60.05%   970.73kB  1.10%  database/sql.driverArgsConnLocked
  886.90kB  1.01% 61.06%   886.90kB  1.01%  github.com/bq2cd/yp-go-metrics/internal/model.NewMetricKeySet (inline)
  879.30kB  1.00% 62.06% 14708.27kB 16.73%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).GetMulti
  858.51kB  0.98% 63.04%   858.51kB  0.98%  github.com/jackc/pgx/v5/internal/iobufpool.init.0.func1
  804.67kB  0.92% 63.95%   804.67kB  0.92%  github.com/huandu/go-sqlbuilder.(*valueStore).Add (inline)
  775.52kB  0.88% 64.83%  1539.84kB  1.75%  github.com/huandu/go-sqlbuilder.(*InsertBuilder).Values
  758.62kB  0.86% 65.70%   982.36kB  1.12%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).storeMetrics
  738.88kB  0.84% 66.54% 10590.48kB 12.05%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiForType
     719kB  0.82% 67.36%      719kB  0.82%  slices.SortedStableFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }].Collect[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }].AppendSeq[go.shape.[]go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" },go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }]-range1
     707kB   0.8% 68.16%  1353.92kB  1.54%  io.ReadAll
     704kB   0.8% 68.96%      704kB   0.8%  compress/flate.(*dictDecoder).init (inline)
  666.38kB  0.76% 69.72%   666.38kB  0.76%  go.uber.org/zap.(*Logger).clone (inline)
  663.34kB  0.75% 70.47%   691.27kB  0.79%  net/textproto.readMIMEHeader
  641.94kB  0.73% 71.20% 11342.24kB 12.90%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByType
  630.21kB  0.72% 71.92%  6151.21kB  7.00%  github.com/bq2cd/yp-go-metrics/pkg/log.(*zapLogger).with
  620.43kB  0.71% 72.63%   620.43kB  0.71%  reflect.growslice
     616kB   0.7% 73.33%   616.06kB   0.7%  github.com/goccy/go-json/internal/decoder.initDecoder.func1
  604.03kB  0.69% 74.01%   746.16kB  0.85%  github.com/huandu/go-sqlbuilder.newSelectBuilder
     595kB  0.68% 74.69%      595kB  0.68%  github.com/goccy/go-json/internal/decoder.NewStream (inline)
     595kB  0.68% 75.37%      595kB  0.68%  net/http.(*Request).WithContext (inline)
     561kB  0.64% 76.01%      561kB  0.64%  reflect.unsafe_NewArray
     476kB  0.54% 76.55%  1375.80kB  1.57%  github.com/goccy/go-json/internal/decoder.(*sliceDecoder).DecodeStream
     475kB  0.54% 77.09%      475kB  0.54%  crypto/internal/fips140/sha256.New (inline)
  429.41kB  0.49% 77.58%  8214.95kB  9.35%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].Select
  415.67kB  0.47% 78.05%   890.72kB  1.01%  crypto/internal/fips140/hmac.New[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }]
  376.06kB  0.43% 78.48%   963.04kB  1.10%  github.com/bq2cd/yp-go-metrics/internal/repository/auditsink.(*fileSink).WriteEvent
     357kB  0.41% 78.88%  1724.12kB  1.96%  net/http.(*conn).readRequest
  333.19kB  0.38% 79.26%  4854.62kB  5.52%  go.uber.org/zap/zapcore.(*sampler).With
  331.87kB  0.38% 79.64%  1812.73kB  2.06%  github.com/jackc/pgx/v5/pgconn.connectOne
  297.50kB  0.34% 79.98%  1239.11kB  1.41%  net/http.readRequest
  286.33kB  0.33% 80.30%   525.47kB   0.6%  context.withCancel (inline)
  264.44kB   0.3% 80.61%     1163kB  1.32%  github.com/jackc/pgx/v5/stdlib.(*Conn).QueryContext
  257.69kB  0.29% 80.90%  2390.29kB  2.72%  github.com/huandu/go-sqlbuilder.(*Args).CompileWithFlavor
  249.89kB  0.28% 81.18%  4521.39kB  5.14%  go.uber.org/zap/zapcore.(*ioCore).clone (inline)
  237.08kB  0.27% 81.45%  2018.54kB  2.30%  database/sql.(*DB).queryDC
     237kB  0.27% 81.72%  2226.52kB  2.53%  github.com/huandu/go-sqlbuilder.(*InsertBuilder).BuildWithFlavor
  218.53kB  0.25% 81.97%   838.96kB  0.95%  reflect.Value.extendSlice
  188.41kB  0.21% 82.19%   542.54kB  0.62%  github.com/goccy/go-json.MarshalContext (inline)
  180.94kB  0.21% 82.39%  2108.88kB  2.40%  github.com/bq2cd/yp-go-metrics/pkg/log.(*zapLogger).log
  167.19kB  0.19% 82.58%  2246.34kB  2.56%  github.com/jackc/pgx/v5.connect
  167.06kB  0.19% 82.77%   596.14kB  0.68%  database/sql.(*DB).beginDC
  165.69kB  0.19% 82.96%   869.69kB  0.99%  compress/flate.NewReader
  157.64kB  0.18% 83.14%   495.17kB  0.56%  github.com/jackc/pgx/v5/stdlib.(*Rows).Next
     154kB  0.18% 83.31%  1633.28kB  1.86%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.NewRetrier[go.shape.[]github.com/bq2cd/yp-go-metrics/internal/model.Metric]
  148.75kB  0.17% 83.48% 47087.37kB 53.57%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*loggerMiddleware).Intercept
  142.11kB  0.16% 83.65%   631.27kB  0.72%  github.com/jackc/pgx/v5/pgconn.ParseConfigWithOptions
  120.38kB  0.14% 83.78%   996.27kB  1.13%  github.com/jackc/pgx/v5/pgproto3.NewFrontend
  118.75kB  0.14% 83.92%  1270.66kB  1.45%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.NewRetrier[go.shape.interface {}]
  118.50kB  0.13% 84.05%   495.62kB  0.56%  github.com/bq2cd/yp-go-metrics/internal/model.NewAuditEvent
  118.44kB  0.13% 84.19%  4977.67kB  5.66%  github.com/jmoiron/sqlx.(*DB).QueryxContext
  114.95kB  0.13% 84.32%  2308.25kB  2.63%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].ConvertMetrics
  114.80kB  0.13% 84.45%  5216.21kB  5.93%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].Insert
   83.30kB 0.095% 84.54%  6234.51kB  7.09%  github.com/bq2cd/yp-go-metrics/pkg/log.(*baseLogger).With
   64.35kB 0.073% 84.62%  1562.62kB  1.78%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value int64 "db:\"value\"" }].Select
   63.30kB 0.072% 84.69%   867.98kB  0.99%  github.com/huandu/go-sqlbuilder.(*Args).add
   57.75kB 0.066% 84.75% 13021.33kB 14.81%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.[]github.com/bq2cd/yp-go-metrics/internal/model.Metric]).Do
   57.27kB 0.065% 84.82%   582.73kB  0.66%  context.WithCancel
   55.55kB 0.063% 84.88%   858.12kB  0.98%  github.com/jmoiron/sqlx.(*DB).BeginTxx
   52.06kB 0.059% 84.94%  4949.36kB  5.63%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.(*ReaderPool).Get
   47.45kB 0.054% 85.00%   542.66kB  0.62%  database/sql.(*Rows).nextLocked
   47.38kB 0.054% 85.05%  1844.05kB  2.10%  github.com/jmoiron/sqlx.scanAll
   46.81kB 0.053% 85.10%  3077.40kB  3.50%  github.com/jackc/pgx/v5/stdlib.(*driverConnector).Connect
   45.09kB 0.051% 85.15%   764.09kB  0.87%  slices.AppendSeq[go.shape.[]go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" },go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }] (inline)
   44.53kB 0.051% 85.20% 21081.22kB 23.98%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.interface {}]).Do
   44.44kB 0.051% 85.26%   586.98kB  0.67%  github.com/bq2cd/yp-go-metrics/internal/repository/auditsink.NewFileSink.func1
   37.27kB 0.042% 85.30%  1104.86kB  1.26%  database/sql.(*DB).execDC
   33.44kB 0.038% 85.34%   664.70kB  0.76%  github.com/jackc/pgx/v5.ParseConfigWithOptions
   27.80kB 0.032% 85.37%   791.89kB   0.9%  slices.Collect[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }] (inline)
   23.72kB 0.027% 85.39%  5046.43kB  5.74%  database/sql.(*DB).query
   22.36kB 0.025% 85.42%  2630.38kB  2.99%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).validateMetrics
   22.27kB 0.025% 85.44%   841.78kB  0.96%  github.com/bq2cd/yp-go-metrics/internal/handler.(*defaultMetricBatchJSONResponder).WriteResponse
   18.58kB 0.021% 85.47%   803.96kB  0.91%  database/sql.(*DB).begin
   17.03kB 0.019% 85.49%   808.42kB  0.92%  github.com/jackc/pgx/v5.(*Conn).Query
   15.12kB 0.017% 85.50%   974.88kB  1.11%  compress/gzip.NewReader (inline)
   15.06kB 0.017% 85.52% 55886.48kB 63.58%  net/http.(*conn).serve
   14.88kB 0.017% 85.54% 53565.79kB 60.94%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*requestIDMiddleware).Intercept
   13.38kB 0.015% 85.55%  3222.48kB  3.67%  database/sql.(*DB).conn
    5.06kB 0.0058% 85.56% 19379.54kB 22.05%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).StartProcessing
    5.02kB 0.0057% 85.56%   875.90kB     1%  github.com/jackc/pgx/v5/pgproto3.newChunkReader (inline)
    3.67kB 0.0042% 85.57%   968.21kB  1.10%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).dumpMetrics
    3.34kB 0.0038% 85.57%  1816.19kB  2.07%  github.com/jackc/pgx/v5/pgconn.connectPreferred
    0.38kB 0.00043% 85.57%   969.09kB  1.10%  github.com/bq2cd/yp-go-metrics/pkg/periodictask.(*timerTask).Run
    0.27kB 0.0003% 85.57%  8197.29kB  9.33%  github.com/pressly/goose/v3/internal/sqlparser.ParseSQLMigration
    0.25kB 0.00028% 85.57%  1048.08kB  1.19%  bytes.(*Buffer).grow
    0.11kB 0.00012% 85.57%  8286.13kB  9.43%  github.com/pressly/goose/v3.(*Migration).run
    0.09kB 0.00011% 85.57%   884.70kB  1.01%  compress/gzip.(*Reader).readHeader
    0.07kB 8e-05% 85.57%   788.23kB   0.9%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.NewReaderPool
    0.06kB 7.1e-05% 85.57%  8488.15kB  9.66%  github.com/bq2cd/yp-go-metrics/internal/app/server.applyMigrations
    0.05kB 5.3e-05% 85.57%   475.05kB  0.54%  crypto/hmac.New.UnwrapNew[go.shape.interface { BlockSize int; Reset; Size int; Sum []uint8; Write  }].func1
    0.05kB 5.3e-05% 85.57%  5737.64kB  6.53%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorMiddleware).decompressRequest
    0.05kB 5.3e-05% 85.57%  2009.31kB  2.29%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).validateRequest
    0.05kB 5.3e-05% 85.57%  4521.44kB  5.14%  go.uber.org/zap/zapcore.(*ioCore).With
    0.05kB 5.3e-05% 85.57%   521.62kB  0.59%  io.copyBuffer
    0.04kB 4.4e-05% 85.57%  8479.96kB  9.65%  github.com/pressly/goose/v3.UpToContext
    0.03kB 3.6e-05% 85.57%  9745.23kB 11.09%  github.com/bq2cd/yp-go-metrics/internal/app/cli.App[go.shape.struct { ListenAddress string; ShutdownTimeout time.Duration; MetricStoreInterval time.Duration; MetricStoreFilePath string; MetricStoreLoadOnStartup bool; DatabaseURL net/url.URL; HMACSecretKey []uint8; AuditFilePath string; AuditURL net/url.URL }].Run
    0.02kB 2.7e-05% 85.57%   788.16kB   0.9%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.newNoopReader
    0.02kB 1.8e-05% 85.57%   969.15kB  1.10%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).launchMetricDumper.func1
         0     0% 85.57%  4003.69kB  4.55%  bufio.NewReader (inline)
         0     0% 85.57%  1047.52kB  1.19%  bytes.(*Buffer).Write
         0     0% 85.57%  4872.02kB  5.54%  compress/gzip.(*Reader).Reset
         0     0% 85.57% 20709.81kB 23.56%  compress/gzip.(*Writer).Write
         0     0% 85.57%   890.72kB  1.01%  crypto/hmac.New
         0     0% 85.57%      475kB  0.54%  crypto/sha256.New
         0     0% 85.57%   803.96kB  0.91%  database/sql.(*DB).BeginTx
         0     0% 85.57%   803.96kB  0.91%  database/sql.(*DB).BeginTx.func1
         0     0% 85.57%  5046.43kB  5.74%  database/sql.(*DB).QueryContext
         0     0% 85.57%  5046.43kB  5.74%  database/sql.(*DB).QueryContext.func1
         0     0% 85.57%  1067.59kB  1.21%  database/sql.(*DB).execDC.func2
         0     0% 85.57%  1485.17kB  1.69%  database/sql.(*DB).queryDC.func1
         0     0% 85.57%  5879.45kB  6.69%  database/sql.(*DB).retry
         0     0% 85.57%   617.52kB   0.7%  database/sql.(*Rows).Next
         0     0% 85.57%   542.66kB  0.62%  database/sql.(*Rows).Next.func1
         0     0% 85.57%  1123.47kB  1.28%  database/sql.(*Tx).ExecContext
         0     0% 85.57%     1163kB  1.32%  database/sql.ctxDriverQuery
         0     0% 85.57%  3329.72kB  3.79%  database/sql.withLock
         0     0% 85.57%  1152.04kB  1.31%  github.com/bq2cd/yp-go-metrics/internal/app/cli.(*profiler).MaybeStartProfiling
         0     0% 85.57%     1152kB  1.31%  github.com/bq2cd/yp-go-metrics/internal/app/cli.(*profiler).maybeStartCPUProfiling
         0     0% 85.57%  8592.32kB  9.77%  github.com/bq2cd/yp-go-metrics/internal/app/cli.App[go.shape.struct { ListenAddress string; ShutdownTimeout time.Duration; MetricStoreInterval time.Duration; MetricStoreFilePath string; MetricStoreLoadOnStartup bool; DatabaseURL net/url.URL; HMACSecretKey []uint8; AuditFilePath string; AuditURL net/url.URL }].run
         0     0% 85.57%   968.21kB  1.10%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).createPeriodicTask.func2
         0     0% 85.57% 19379.54kB 22.05%  github.com/bq2cd/yp-go-metrics/internal/app/server.(*server).launchBatchWriter.func1
         0     0% 85.57%  8592.32kB  9.77%  github.com/bq2cd/yp-go-metrics/internal/app/server.Run
         0     0% 85.57%  8492.45kB  9.66%  github.com/bq2cd/yp-go-metrics/internal/app/server.applyMigrationsWithRetries
         0     0% 85.57%  8488.15kB  9.66%  github.com/bq2cd/yp-go-metrics/internal/app/server.applyMigrationsWithRetries.func1
         0     0% 85.57%  8522.61kB  9.70%  github.com/bq2cd/yp-go-metrics/internal/app/server.initStorage
         0     0% 85.57% 20539.51kB 23.37%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).ServeHTTP
         0     0% 85.57%   841.78kB  0.96%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).respondOK
         0     0% 85.57% 14065.52kB 16.00%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).retrieveMetrics
         0     0% 85.57% 53134.30kB 60.45%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorMiddleware).Intercept
         0     0% 85.57%   788.23kB   0.9%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorMiddleware).decompressRequest.func1
         0     0% 85.57% 19921.88kB 22.66%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorResponseWriter).Write
         0     0% 85.57% 43597.02kB 49.60%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).Intercept
         0     0% 85.57%   622.12kB  0.71%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).signResponse
         0     0% 85.57% 20973.56kB 23.86%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).writeResponse
         0     0% 85.57%   443.31kB   0.5%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerResponseWriter).Write
         0     0% 85.57% 19921.88kB 22.66%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*loggerResponseWriter).Write
         0     0% 85.57% 53565.79kB 60.94%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*middlewareHandler).ServeHTTP
         0     0% 85.57% 20539.77kB 23.37%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*recovererMiddleware).Intercept
         0     0% 85.57% 53932.46kB 61.36%  github.com/bq2cd/yp-go-metrics/internal/handler/router.(*Router).ServeHTTP
         0     0% 85.57%  1523.84kB  1.73%  github.com/bq2cd/yp-go-metrics/internal/model.NewMetricSet (inline)
         0     0% 85.57%   825.65kB  0.94%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).GetAll
         0     0% 85.57% 13859.43kB 15.77%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).SetMulti
         0     0% 85.57% 14654.61kB 16.67%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByTypeWithRetries
         0     0% 85.57% 11342.24kB 12.90%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).getMultiByTypeWithRetries.func1
         0     0% 85.57% 11262.52kB 12.81%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMulti
         0     0% 85.57%  8811.97kB 10.02%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiForType
         0     0% 85.57% 13859.43kB 15.77%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiWithRetries
         0     0% 85.57% 11262.52kB 12.81%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.(*sqlStorage).setMultiWithRetries.func1
         0     0% 85.57%   711.59kB  0.81%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].ConvertMetrics.Keys[go.shape.map[github.com/bq2cd/yp-go-metrics/internal/model.MetricKey]github.com/bq2cd/yp-go-metrics/internal/model.Metric,go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" },go.shape.struct { ID string "json:\"id\""; Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; Delta *int64 "json:\"delta,omitempty\""; Value *float64 "json:\"value,omitempty\""; Hash github.com/bq2cd/yp-go-metrics/internal/model.MetricHash "json:\"hash,omitempty\"" }].func2
         0     0% 85.57%  1251.69kB  1.42%  github.com/bq2cd/yp-go-metrics/internal/repository/sqlstorage.sqlHandlerImpl[go.shape.struct { ID string "db:\"metric_id\""; Value float64 "db:\"value\"" }].ConvertMetrics.func1
         0     0% 85.57%   963.04kB  1.10%  github.com/bq2cd/yp-go-metrics/internal/service.(*auditEventProcessor).processEvent.func1
         0     0% 85.57%   495.62kB  0.56%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricAuditor).RecordMetricsUploaded
         0     0% 85.57%   869.12kB  0.99%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricSnapshotter).DumpClose
         0     0% 85.57%   825.65kB  0.94%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricStorer).RetrieveAll
         0     0% 85.57% 14065.52kB 16.00%  github.com/bq2cd/yp-go-metrics/internal/service.(*metricStorer).RetrieveBatch
         0     0% 85.57%  5515.05kB  6.27%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).accumulateCounters
         0     0% 85.57% 19374.48kB 22.04%  github.com/bq2cd/yp-go-metrics/internal/service.(*storageBatchWriter).processBatchTx
         0     0% 85.57% 19921.88kB 22.66%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.(*Writer).Write
         0     0% 85.57%   974.88kB  1.11%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.NewReaderPool.func1
         0     0% 85.57%      474kB  0.54%  github.com/bq2cd/yp-go-metrics/pkg/hmacsigner.(*hmacSigner).Sign
         0     0% 85.57%   476.09kB  0.54%  github.com/bq2cd/yp-go-metrics/pkg/hmacsigner.(*hmacSigner).Verify
         0     0% 85.57%  2108.88kB  2.40%  github.com/bq2cd/yp-go-metrics/pkg/log.(*eventBuilder).Msg
         0     0% 85.57% 53932.46kB 61.36%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 85.57% 20539.77kB 23.37%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 85.57%  2013.02kB  2.29%  github.com/goccy/go-json.(*Decoder).Decode (inline)
         0     0% 85.57%  2013.02kB  2.29%  github.com/goccy/go-json.(*Decoder).DecodeWithOption
         0     0% 85.57%   521.20kB  0.59%  github.com/goccy/go-json.(*Encoder).Encode (inline)
         0     0% 85.57%   521.20kB  0.59%  github.com/goccy/go-json.(*Encoder).EncodeWithOption
         0     0% 85.57%      498kB  0.57%  github.com/goccy/go-json.(*Encoder).encodeWithOption
         0     0% 85.57%      595kB  0.68%  github.com/goccy/go-json.NewDecoder (inline)
         0     0% 85.57%   637.23kB  0.72%  github.com/goccy/go-json/internal/decoder.CompileToGetDecoder
         0     0% 85.57%   616.16kB   0.7%  github.com/goccy/go-json/internal/decoder.initDecoder
         0     0% 85.57%   867.98kB  0.99%  github.com/huandu/go-sqlbuilder.(*Args).Add
         0     0% 85.57%  1882.72kB  2.14%  github.com/huandu/go-sqlbuilder.(*Args).compileDigits
         0     0% 85.57%  1882.72kB  2.14%  github.com/huandu/go-sqlbuilder.(*Args).compileSuccessive
         0     0% 85.57%   476.53kB  0.54%  github.com/huandu/go-sqlbuilder.(*Cond).In.func1
         0     0% 85.57%  1247.86kB  1.42%  github.com/huandu/go-sqlbuilder.(*SelectBuilder).BuildWithFlavor
         0     0% 85.57%   706.08kB   0.8%  github.com/huandu/go-sqlbuilder.(*WhereClause).BuildWithFlavor
         0     0% 85.57%   626.73kB  0.71%  github.com/huandu/go-sqlbuilder.(*clause).Build
         0     0% 85.57%  1236.50kB  1.41%  github.com/huandu/go-sqlbuilder.(*stringBuilder).WriteString (inline)
         0     0% 85.57%   629.28kB  0.72%  github.com/huandu/go-sqlbuilder.(*stringBuilder).WriteStrings
         0     0% 85.57%   746.16kB  0.85%  github.com/huandu/go-sqlbuilder.Flavor.NewSelectBuilder (inline)
         0     0% 85.57%   746.16kB  0.85%  github.com/huandu/go-sqlbuilder.NewSelectBuilder
         0     0% 85.57%  2365.88kB  2.69%  github.com/jackc/pgx/v5.ConnectConfig
         0     0% 85.57%   664.70kB  0.76%  github.com/jackc/pgx/v5.ParseConfig (inline)
         0     0% 85.57%   870.88kB  0.99%  github.com/jackc/pgx/v5/internal/iobufpool.Get
         0     0% 85.57%  1851.04kB  2.11%  github.com/jackc/pgx/v5/pgconn.ConnectConfig
         0     0% 85.57%   996.27kB  1.13%  github.com/jackc/pgx/v5/pgconn.ParseConfigWithOptions.func1
         0     0% 85.57%  6845.41kB  7.79%  github.com/jmoiron/sqlx.(*DB).SelectContext
         0     0% 85.57%  6845.41kB  7.79%  github.com/jmoiron/sqlx.SelectContext
         0     0% 85.57%  8286.13kB  9.43%  github.com/pressly/goose/v3.(*Migration).UpContext (inline)
         0     0% 85.57%  8479.96kB  9.65%  github.com/pressly/goose/v3.UpContext (inline)
         0     0% 85.57%  1927.95kB  2.19%  go.uber.org/zap.(*Logger).Log
         0     0% 85.57%     5521kB  6.28%  go.uber.org/zap.(*Logger).With
         0     0% 85.57%  5522.31kB  6.28%  go.uber.org/zap/buffer.Pool.Get
         0     0% 85.57%     5511kB  6.27%  go.uber.org/zap/internal/bufferpool.init.NewPool.New[go.shape.*uint8].func2
         0     0% 85.57%  5904.60kB  6.72%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0% 85.57%  1655.32kB  1.88%  go.uber.org/zap/zapcore.(*CheckedEntry).Write
         0     0% 85.57%  1653.04kB  1.88%  go.uber.org/zap/zapcore.(*ioCore).Write
         0     0% 85.57%  4271.50kB  4.86%  go.uber.org/zap/zapcore.(*jsonEncoder).Clone
         0     0% 85.57%  1642.21kB  1.87%  go.uber.org/zap/zapcore.(*jsonEncoder).EncodeEntry
         0     0% 85.57%  4958.30kB  5.64%  go.uber.org/zap/zapcore.(*jsonEncoder).clone
         0     0% 85.57%   935.30kB  1.06%  go.uber.org/zap/zapcore.EntryCaller.TrimmedPath
         0     0% 85.57%   935.30kB  1.06%  go.uber.org/zap/zapcore.ShortCallerEncoder
         0     0% 85.57%   963.04kB  1.10%  golang.org/x/sync/errgroup.(*Group).Go.func1
         0     0% 85.57%   646.92kB  0.74%  io.(*teeReader).Read
         0     0% 85.57%   521.62kB  0.59%  io.Copy (inline)
         0     0% 85.57%  9745.23kB 11.09%  main.main
         0     0% 85.57%  9745.23kB 11.09%  main.run
         0     0% 85.57% 20539.77kB 23.37%  net/http.HandlerFunc.ServeHTTP
         0     0% 85.57% 53932.46kB 61.36%  net/http.serverHandler.ServeHTTP
         0     0% 85.57%   691.27kB  0.79%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0% 85.57%   838.96kB  0.95%  reflect.Append
         0     0% 85.57%   620.43kB  0.71%  reflect.Value.grow
         0     0% 85.57%  9745.23kB 11.09%  runtime.main
         0     0% 85.57%  1251.69kB  1.42%  slices.SortStableFunc[go.shape.[]go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" },go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }] (inline)
         0     0% 85.57%  2043.58kB  2.32%  slices.SortedStableFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }]
         0     0% 85.57%  1251.69kB  1.42%  slices.insertionSortCmpFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }] (inline)
         0     0% 85.57%  1251.69kB  1.42%  slices.stableCmpFunc[go.shape.struct { Type github.com/bq2cd/yp-go-metrics/internal/model.MetricType "json:\"type\""; ID string "json:\"id\"" }]
         0     0% 85.57%  2006.67kB  2.28%  sync.(*Once).Do (inline)
         0     0% 85.57%  2006.67kB  2.28%  sync.(*Once).doSlow
         0     0% 85.57% 16204.82kB 18.44%  sync.(*Pool).Get
```
