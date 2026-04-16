# Memory optimization results

## 2026-04-16

### Agent

Diff: `go tool pprof -top -sample_index=alloc_space -diff_base=base/agent-mem.pprof result/agent-mem.pprof`

```
File: agent-3561658577
Type: alloc_space
Time: 2026-04-16 11:19:40 MSK
Showing nodes accounting for -1366111.47kB, 88.79% of 1538536.52kB total
Dropped 381 nodes (cum <= 7692.68kB)
      flat  flat%   sum%        cum   cum%
-751032.42kB 48.81% 48.81% -1358203.59kB 88.28%  compress/flate.NewWriter (inline)
-370879.95kB 24.11% 72.92% -607171.17kB 39.46%  compress/flate.(*compressor).init
 -231800kB 15.07% 87.99%  -231800kB 15.07%  compress/flate.newDeflateFast (inline)
33920.20kB  2.20% 85.78% -1324354.16kB 86.08%  io.copyBuffer
  -20228kB  1.31% 87.10%   -20228kB  1.31%  regexp.(*bitState).reset
  -18752kB  1.22% 88.32%   -18752kB  1.22%  net/http.init.func15
-7175.34kB  0.47% 88.78% -7175.34kB  0.47%  compress/flate.(*huffmanEncoder).generate
 -135.11kB 0.0088% 88.79% -1364359.16kB 88.68%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).SendBatch
  -55.88kB 0.0036% 88.79% -1405126.59kB 91.33%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).setBody
   49.69kB 0.0032% 88.79% 74005.33kB  4.81%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).prepareBody
  -10.47kB 0.00068% 88.79% -27198.82kB  1.77%  github.com/go-resty/resty/v2.(*Client).execute
   -8.25kB 0.00054% 88.79% -20239.72kB  1.32%  net/http.(*Request).write
   -6.19kB 0.0004% 88.79% -1433451.85kB 93.17%  github.com/bq2cd/yp-go-metrics/pkg/retrymgr.(*retrier[go.shape.*uint8]).Do
    3.56kB 0.00023% 88.79% -1364957.48kB 88.72%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).reportWorker (inline)
   -1.31kB 8.5e-05% 88.79% -20241.03kB  1.32%  net/http.(*persistConn).writeLoop
         0     0% 88.79% -1396950.20kB 90.80%  bytes.(*Reader).WriteTo
         0     0% 88.79% -7504.19kB  0.49%  compress/flate.(*Writer).Close (inline)
         0     0% 88.79% -7504.19kB  0.49%  compress/flate.(*compressor).close
         0     0% 88.79% -7175.34kB  0.47%  compress/flate.(*compressor).encSpeed
         0     0% 88.79% -7667.41kB   0.5%  compress/gzip.(*Writer).Close
         0     0% 88.79% -1358274.41kB 88.28%  compress/gzip.(*Writer).Write
         0     0% 88.79% -1364957.48kB 88.72%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).processBatches.func1
         0     0% 88.79% -1364961.05kB 88.72%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*reporter).reportBatch
         0     0% 88.79% 72991.02kB  4.74%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).compressBody
         0     0% 88.79% -1433205.51kB 93.15%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendSingleRequest
         0     0% 88.79% -1360361.02kB 88.42%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendWithRetries
         0     0% 88.79% -1433205.51kB 93.15%  github.com/bq2cd/yp-go-metrics/internal/app/agent.(*senderJSON).sendWithRetries.func1
         0     0% 88.79% 38675.80kB  2.51%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.(*Writer).Write
         0     0% 88.79% -13846.26kB   0.9%  github.com/go-resty/resty/v2.(*Client).executeBefore
         0     0% 88.79% -27198.82kB  1.77%  github.com/go-resty/resty/v2.(*Request).Execute
         0     0% 88.79% -27198.82kB  1.77%  github.com/go-resty/resty/v2.(*Request).Send (inline)
         0     0% 88.79% -20709.93kB  1.35%  github.com/go-resty/resty/v2.IsJSONType (inline)
         0     0% 88.79% -13698.80kB  0.89%  github.com/go-resty/resty/v2.parseRequestHeader
         0     0% 88.79% -7015.25kB  0.46%  github.com/go-resty/resty/v2.parseResponseBody
         0     0% 88.79% -1365821.48kB 88.77%  golang.org/x/sync/errgroup.(*Group).Go.func1
         0     0% 88.79% -1324354.36kB 86.08%  io.Copy (inline)
         0     0% 88.79% -19100.78kB  1.24%  net/http.(*transferWriter).doBodyCopy
         0     0% 88.79% -19103.92kB  1.24%  net/http.(*transferWriter).writeBody
         0     0% 88.79% -19097.22kB  1.24%  net/http.getCopyBuf (inline)
         0     0% 88.79% -20705.93kB  1.35%  regexp.(*Regexp).MatchString (inline)
         0     0% 88.79% -20705.93kB  1.35%  regexp.(*Regexp).backtrack
         0     0% 88.79% -20705.93kB  1.35%  regexp.(*Regexp).doExecute
         0     0% 88.79% -20705.93kB  1.35%  regexp.(*Regexp).doMatch (inline)
         0     0% 88.79% -24048.24kB  1.56%  sync.(*Pool).Get
```

### Server

Diff: `go tool pprof -top -sample_index=alloc_space -diff_base=base/server-mem.pprof result/server-mem.pprof`

```
File: server-222229620
Type: alloc_space
Time: 2026-04-16 11:19:40 MSK
Showing nodes accounting for -1415348.98kB, 91.26% of 1550966.55kB total
Dropped 788 nodes (cum <= 7754.83kB)
      flat  flat%   sum%        cum   cum%
 -758808kB 48.92% 48.92% -1372649.62kB 88.50%  compress/flate.NewWriter (inline)
 -375040kB 24.18% 73.11% -613841.62kB 39.58%  compress/flate.(*compressor).init
 -234400kB 15.11% 88.22%  -234400kB 15.11%  compress/flate.newDeflateFast (inline)
  -37440kB  2.41% 90.63%   -37440kB  2.41%  compress/flate.(*dictDecoder).init (inline)
-8811.56kB  0.57% 91.20% -46251.61kB  2.98%  compress/flate.NewReader
 -804.38kB 0.052% 91.25% -51864.44kB  3.34%  compress/gzip.NewReader (inline)
  -37.50kB 0.0024% 91.26% -1388304.16kB 89.51%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*loggerMiddleware).Intercept
   -3.75kB 0.00024% 91.26% -1442918.44kB 93.03%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*requestIDMiddleware).Intercept
   -3.75kB 0.00024% 91.26% -1454195.33kB 93.76%  net/http.(*conn).serve
   -0.05kB 3e-06% 91.26% -47147.80kB  3.04%  compress/gzip.(*Reader).Reset
         0     0% 91.26% -7920.23kB  0.51%  bufio.(*Writer).Flush
         0     0% 91.26% -46255.31kB  2.98%  compress/gzip.(*Reader).readHeader
         0     0% 91.26% -1372649.56kB 88.50%  compress/gzip.(*Writer).Write
         0     0% 91.26% -10218.12kB  0.66%  github.com/bq2cd/yp-go-metrics/internal/handler.(*updateBatchJSONHandler).ServeHTTP
         0     0% 91.26% -1442809.62kB 93.03%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorMiddleware).Intercept
         0     0% 91.26% -47101.72kB  3.04%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorMiddleware).decompressRequest
         0     0% 91.26% -1373437.50kB 88.55%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*compressorResponseWriter).Write
         0     0% 91.26% -1384340.12kB 89.26%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).Intercept
         0     0% 91.26% -1373704.86kB 88.57%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*hmacSignerMiddleware).writeResponse
         0     0% 91.26% -1373437.50kB 88.55%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*loggerResponseWriter).Write
         0     0% 91.26% -1442918.44kB 93.03%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*middlewareHandler).ServeHTTP
         0     0% 91.26% -10226.48kB  0.66%  github.com/bq2cd/yp-go-metrics/internal/handler/middleware.(*recovererMiddleware).Intercept
         0     0% 91.26% -1443732.41kB 93.09%  github.com/bq2cd/yp-go-metrics/internal/handler/router.(*Router).ServeHTTP
         0     0% 91.26% 19921.88kB  1.28%  github.com/bq2cd/yp-go-metrics/pkg/gzippool.(*Writer).Write
         0     0% 91.26% -1443732.41kB 93.09%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 91.26% -10226.48kB  0.66%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 91.26% -7920.23kB  0.51%  net/http.(*chunkWriter).Write
         0     0% 91.26% -7920.23kB  0.51%  net/http.(*chunkWriter).writeHeader
         0     0% 91.26% -8485.59kB  0.55%  net/http.(*response).finishRequest
         0     0% 91.26% -10226.48kB  0.66%  net/http.HandlerFunc.ServeHTTP
         0     0% 91.26% -1443732.41kB 93.09%  net/http.serverHandler.ServeHTTP
         0     0% 91.26% -17951.83kB  1.16%  sync.(*Pool).Get
```
