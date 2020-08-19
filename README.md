# Reproduce issue

To reproduce run the test PSK server with multiple clients in parallel.

- Start server.
  ```shell
  $ go run ./examples/listen/psk/
  ```
- Start multiple clients.
  ```shell
  mkdir -p build && go build -o build/ ./examples/dial/psk && for i in {1..10}; do ./build/psk <<< 'exit' &; done
  ```

The outputs should that the client are processing the handshakes intertwined, so the logs from one identifier are mixed with the other ones. On the server the logs from one identifier are in a block. So the full handshake for one identifier is logged before the handshake for the next identifier starts.

See [pion/dtls#279](https://github.com/pion/dtls/issues/279) for reference to this issue.

As noted in a comment on that issue, a workaround for the moment is to spawn multiple go routines. The [patch](https://gist.github.com/Sean-Der/a3a6abbe9b78ced613590312a18dc8a4), with the random hanging in `conn.go` removed is applied here as test.

<details><summary>Patch for concurrent listeners.</summary>

```diff
diff --git a/examples/listen/psk/main.go b/examples/listen/psk/main.go
index 72a6c23..06b444d 100644
--- a/examples/listen/psk/main.go
+++ b/examples/listen/psk/main.go
@@ -49,23 +49,25 @@ func main() {
 	// Simulate a chat session
 	hub := util.NewHub()

-	go func() {
-		for {
-			// Wait for a connection.
-			conn, err := listener.Accept()
-			util.Check(err)
-			// defer conn.Close() // TODO: graceful shutdown
+	for i := 0; i <= 5; i++ {
+		go func() {
+			for {
+				// Wait for a connection.
+				conn, err := listener.Accept()
+				util.Check(err)
+				// defer conn.Close() // TODO: graceful shutdown

-			// `conn` is of type `net.Conn` but may be casted to `dtls.Conn`
-			// using `dtlsConn := conn.(*dtls.Conn)` in order to to expose
-			// functions like `ConnectionState` etc.
+				// `conn` is of type `net.Conn` but may be casted to `dtls.Conn`
+				// using `dtlsConn := conn.(*dtls.Conn)` in order to to expose
+				// functions like `ConnectionState` etc.

-			// Register the connection with the chat hub
-			if err == nil {
-				hub.Register(conn)
+				// Register the connection with the chat hub
+				if err == nil {
+					hub.Register(conn)
+				}
 			}
-		}
-	}()
+		}()
+	}

 	// Start chatting
 	hub.Chat()
```

</details>

- Original situation.

  <details><summary>Server-side log with 1 concurrent listener (original situation).</summary>

  ```shell
  $ go run ./examples/listen/psk/
  Listening
  bMcmYJ dtls TRACE: 07:06:43.403976 handshaker.go:151: [handshake:server] Flight 0: Preparing
  bMcmYJ dtls TRACE: 07:06:43.404029 handshaker.go:151: [handshake:server] Flight 0: Sending
  bMcmYJ dtls TRACE: 07:06:43.404034 handshaker.go:151: [handshake:server] Flight 0: Waiting
  bMcmYJ dtls TRACE: 07:06:43.404090 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  bMcmYJ dtls TRACE: 07:06:43.404095 handshaker.go:151: [handshake:server] Flight 2: Preparing
  bMcmYJ dtls TRACE: 07:06:43.404099 handshaker.go:151: [handshake:server] Flight 2: Sending
  bMcmYJ dtls TRACE: 07:06:43.404105 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  bMcmYJ dtls TRACE: 07:06:43.404131 handshaker.go:151: [handshake:server] Flight 2: Waiting
  bMcmYJ dtls TRACE: 07:06:43.635149 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  bMcmYJ dtls TRACE: 07:06:43.635188 handshaker.go:151: [handshake:server] Flight 4: Preparing
  bMcmYJ dtls TRACE: 07:06:43.635235 handshaker.go:151: [handshake:server] Flight 4: Sending
  bMcmYJ dtls TRACE: 07:06:43.635279 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  bMcmYJ dtls TRACE: 07:06:43.635310 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  bMcmYJ dtls TRACE: 07:06:43.635340 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  bMcmYJ dtls TRACE: 07:06:43.635463 handshaker.go:151: [handshake:server] Flight 4: Waiting
  bMcmYJ dtls DEBUG: 07:06:43.779264 conn.go:722: CipherSuite not initialized, queuing packet
  bMcmYJ dtls DEBUG: 07:06:43.779314 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  bMcmYJ dtls TRACE: 07:06:43.779493 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  bMcmYJ dtls TRACE: 07:06:43.779595 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  bMcmYJ dtls TRACE: 07:06:43.779620 handshaker.go:151: [handshake:server] Flight 6: Preparing
  bMcmYJ dtls TRACE: 07:06:43.779667 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  bMcmYJ dtls TRACE: 07:06:43.779686 handshaker.go:151: [handshake:server] Flight 6: Sending
  bMcmYJ dtls TRACE: 07:06:43.779710 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  bMcmYJ dtls TRACE: 07:06:43.779846 handshaker.go:151: [handshake:server] Flight 6: Finished
  bMcmYJ dtls TRACE: 07:06:43.779879 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:52730
  9cnnY5 dtls TRACE: 07:06:43.780176 handshaker.go:151: [handshake:server] Flight 0: Preparing
  9cnnY5 dtls TRACE: 07:06:43.780213 handshaker.go:151: [handshake:server] Flight 0: Sending
  9cnnY5 dtls TRACE: 07:06:43.780232 handshaker.go:151: [handshake:server] Flight 0: Waiting
  9cnnY5 dtls TRACE: 07:06:43.780451 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  9cnnY5 dtls TRACE: 07:06:43.780470 handshaker.go:151: [handshake:server] Flight 2: Preparing
  9cnnY5 dtls TRACE: 07:06:43.780489 handshaker.go:151: [handshake:server] Flight 2: Sending
  9cnnY5 dtls TRACE: 07:06:43.780510 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  9cnnY5 dtls TRACE: 07:06:43.780592 handshaker.go:151: [handshake:server] Flight 2: Waiting
  9cnnY5 dtls TRACE: 07:06:43.915276 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  9cnnY5 dtls TRACE: 07:06:43.915317 handshaker.go:151: [handshake:server] Flight 4: Preparing
  9cnnY5 dtls TRACE: 07:06:43.915357 handshaker.go:151: [handshake:server] Flight 4: Sending
  9cnnY5 dtls TRACE: 07:06:43.915395 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  9cnnY5 dtls TRACE: 07:06:43.915425 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  9cnnY5 dtls TRACE: 07:06:43.915450 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  9cnnY5 dtls TRACE: 07:06:43.915567 handshaker.go:151: [handshake:server] Flight 4: Waiting
  bMcmYJ dtls TRACE: 07:06:43.929720 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:52730
  9cnnY5 dtls DEBUG: 07:06:43.989875 conn.go:722: CipherSuite not initialized, queuing packet
  9cnnY5 dtls DEBUG: 07:06:43.991016 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  9cnnY5 dtls TRACE: 07:06:43.991483 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  9cnnY5 dtls TRACE: 07:06:43.991954 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  9cnnY5 dtls TRACE: 07:06:43.992008 handshaker.go:151: [handshake:server] Flight 6: Preparing
  9cnnY5 dtls TRACE: 07:06:43.992083 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  9cnnY5 dtls TRACE: 07:06:43.992641 handshaker.go:151: [handshake:server] Flight 6: Sending
  9cnnY5 dtls TRACE: 07:06:43.992676 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  9cnnY5 dtls TRACE: 07:06:43.992928 handshaker.go:151: [handshake:server] Flight 6: Finished
  9cnnY5 dtls TRACE: 07:06:43.993108 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:58313
  K4qHsf dtls TRACE: 07:06:43.993477 handshaker.go:151: [handshake:server] Flight 0: Preparing
  K4qHsf dtls TRACE: 07:06:43.994046 handshaker.go:151: [handshake:server] Flight 0: Sending
  K4qHsf dtls TRACE: 07:06:43.994074 handshaker.go:151: [handshake:server] Flight 0: Waiting
  K4qHsf dtls TRACE: 07:06:43.994263 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  K4qHsf dtls TRACE: 07:06:43.994324 handshaker.go:151: [handshake:server] Flight 2: Preparing
  K4qHsf dtls TRACE: 07:06:43.994346 handshaker.go:151: [handshake:server] Flight 2: Sending
  K4qHsf dtls TRACE: 07:06:43.994369 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  K4qHsf dtls TRACE: 07:06:43.994471 handshaker.go:151: [handshake:server] Flight 2: Waiting
  K4qHsf dtls TRACE: 07:06:44.081257 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  K4qHsf dtls TRACE: 07:06:44.081270 handshaker.go:151: [handshake:server] Flight 4: Preparing
  K4qHsf dtls TRACE: 07:06:44.081309 handshaker.go:151: [handshake:server] Flight 4: Sending
  K4qHsf dtls TRACE: 07:06:44.081317 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  K4qHsf dtls TRACE: 07:06:44.081324 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  K4qHsf dtls TRACE: 07:06:44.081330 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  K4qHsf dtls TRACE: 07:06:44.081364 handshaker.go:151: [handshake:server] Flight 4: Waiting
  9cnnY5 dtls TRACE: 07:06:44.193163 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:58313
  K4qHsf dtls DEBUG: 07:06:44.237871 conn.go:722: CipherSuite not initialized, queuing packet
  K4qHsf dtls DEBUG: 07:06:44.237887 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  K4qHsf dtls TRACE: 07:06:44.237929 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  K4qHsf dtls TRACE: 07:06:44.237945 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  K4qHsf dtls TRACE: 07:06:44.237952 handshaker.go:151: [handshake:server] Flight 6: Preparing
  K4qHsf dtls TRACE: 07:06:44.237963 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  K4qHsf dtls TRACE: 07:06:44.237968 handshaker.go:151: [handshake:server] Flight 6: Sending
  K4qHsf dtls TRACE: 07:06:44.237978 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  K4qHsf dtls TRACE: 07:06:44.238013 handshaker.go:151: [handshake:server] Flight 6: Finished
  K4qHsf dtls TRACE: 07:06:44.238021 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:32888
  HYrUwl dtls TRACE: 07:06:44.238087 handshaker.go:151: [handshake:server] Flight 0: Preparing
  HYrUwl dtls TRACE: 07:06:44.238095 handshaker.go:151: [handshake:server] Flight 0: Sending
  HYrUwl dtls TRACE: 07:06:44.238100 handshaker.go:151: [handshake:server] Flight 0: Waiting
  HYrUwl dtls TRACE: 07:06:44.238149 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  HYrUwl dtls TRACE: 07:06:44.238157 handshaker.go:151: [handshake:server] Flight 2: Preparing
  HYrUwl dtls TRACE: 07:06:44.238162 handshaker.go:151: [handshake:server] Flight 2: Sending
  HYrUwl dtls TRACE: 07:06:44.238167 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  HYrUwl dtls TRACE: 07:06:44.238188 handshaker.go:151: [handshake:server] Flight 2: Waiting
  HYrUwl dtls TRACE: 07:06:44.412056 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  HYrUwl dtls TRACE: 07:06:44.412097 handshaker.go:151: [handshake:server] Flight 4: Preparing
  HYrUwl dtls TRACE: 07:06:44.412127 handshaker.go:151: [handshake:server] Flight 4: Sending
  HYrUwl dtls TRACE: 07:06:44.412154 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  HYrUwl dtls TRACE: 07:06:44.412194 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  HYrUwl dtls TRACE: 07:06:44.412219 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  HYrUwl dtls TRACE: 07:06:44.412313 handshaker.go:151: [handshake:server] Flight 4: Waiting
  K4qHsf dtls TRACE: 07:06:44.435649 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:32888
  HYrUwl dtls DEBUG: 07:06:44.570091 conn.go:722: CipherSuite not initialized, queuing packet
  HYrUwl dtls DEBUG: 07:06:44.570105 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  HYrUwl dtls TRACE: 07:06:44.570156 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  HYrUwl dtls TRACE: 07:06:44.570175 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  HYrUwl dtls TRACE: 07:06:44.570181 handshaker.go:151: [handshake:server] Flight 6: Preparing
  HYrUwl dtls TRACE: 07:06:44.570198 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  HYrUwl dtls TRACE: 07:06:44.570204 handshaker.go:151: [handshake:server] Flight 6: Sending
  HYrUwl dtls TRACE: 07:06:44.570313 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  HYrUwl dtls TRACE: 07:06:44.570371 handshaker.go:151: [handshake:server] Flight 6: Finished
  HYrUwl dtls TRACE: 07:06:44.570496 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:40429
  Vvlb7y dtls TRACE: 07:06:44.570583 handshaker.go:151: [handshake:server] Flight 0: Preparing
  Vvlb7y dtls TRACE: 07:06:44.570621 handshaker.go:151: [handshake:server] Flight 0: Sending
  Vvlb7y dtls TRACE: 07:06:44.570630 handshaker.go:151: [handshake:server] Flight 0: Waiting
  Vvlb7y dtls TRACE: 07:06:44.570697 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  Vvlb7y dtls TRACE: 07:06:44.570734 handshaker.go:151: [handshake:server] Flight 2: Preparing
  Vvlb7y dtls TRACE: 07:06:44.570742 handshaker.go:151: [handshake:server] Flight 2: Sending
  Vvlb7y dtls TRACE: 07:06:44.570753 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  Vvlb7y dtls TRACE: 07:06:44.570784 handshaker.go:151: [handshake:server] Flight 2: Waiting
  Vvlb7y dtls TRACE: 07:06:44.570796 handshaker.go:274: [handshake:server] Flight 2 -> Flight 2
  Vvlb7y dtls TRACE: 07:06:44.570801 handshaker.go:151: [handshake:server] Flight 2: Preparing
  Vvlb7y dtls TRACE: 07:06:44.570805 handshaker.go:151: [handshake:server] Flight 2: Sending
  Vvlb7y dtls TRACE: 07:06:44.570810 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  Vvlb7y dtls TRACE: 07:06:44.570826 handshaker.go:151: [handshake:server] Flight 2: Waiting
  Vvlb7y dtls TRACE: 07:06:44.672264 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  Vvlb7y dtls TRACE: 07:06:44.672317 handshaker.go:151: [handshake:server] Flight 4: Preparing
  Vvlb7y dtls TRACE: 07:06:44.672331 handshaker.go:151: [handshake:server] Flight 4: Sending
  Vvlb7y dtls TRACE: 07:06:44.672356 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  Vvlb7y dtls TRACE: 07:06:44.672377 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  Vvlb7y dtls TRACE: 07:06:44.672385 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  Vvlb7y dtls TRACE: 07:06:44.672418 handshaker.go:151: [handshake:server] Flight 4: Waiting
  HYrUwl dtls TRACE: 07:06:44.753113 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:40429
  Vvlb7y dtls DEBUG: 07:06:44.899440 conn.go:722: CipherSuite not initialized, queuing packet
  Vvlb7y dtls DEBUG: 07:06:44.899489 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  Vvlb7y dtls TRACE: 07:06:44.899594 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  Vvlb7y dtls TRACE: 07:06:44.899637 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  Vvlb7y dtls TRACE: 07:06:44.899658 handshaker.go:151: [handshake:server] Flight 6: Preparing
  Vvlb7y dtls TRACE: 07:06:44.899692 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  Vvlb7y dtls TRACE: 07:06:44.899705 handshaker.go:151: [handshake:server] Flight 6: Sending
  Vvlb7y dtls TRACE: 07:06:44.899721 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  Vvlb7y dtls TRACE: 07:06:44.899811 handshaker.go:151: [handshake:server] Flight 6: Finished
  Vvlb7y dtls TRACE: 07:06:44.899838 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:33136
  Zq0zwo dtls TRACE: 07:06:44.900061 handshaker.go:151: [handshake:server] Flight 0: Preparing
  Zq0zwo dtls TRACE: 07:06:44.900095 handshaker.go:151: [handshake:server] Flight 0: Sending
  Zq0zwo dtls TRACE: 07:06:44.900106 handshaker.go:151: [handshake:server] Flight 0: Waiting
  Zq0zwo dtls TRACE: 07:06:44.901576 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  Zq0zwo dtls TRACE: 07:06:44.901615 handshaker.go:151: [handshake:server] Flight 2: Preparing
  Zq0zwo dtls TRACE: 07:06:44.901631 handshaker.go:151: [handshake:server] Flight 2: Sending
  Zq0zwo dtls TRACE: 07:06:44.901650 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  Zq0zwo dtls TRACE: 07:06:44.901741 handshaker.go:151: [handshake:server] Flight 2: Waiting
  Zq0zwo dtls TRACE: 07:06:44.901791 handshaker.go:274: [handshake:server] Flight 2 -> Flight 2
  Zq0zwo dtls TRACE: 07:06:44.901807 handshaker.go:151: [handshake:server] Flight 2: Preparing
  Zq0zwo dtls TRACE: 07:06:44.901820 handshaker.go:151: [handshake:server] Flight 2: Sending
  Zq0zwo dtls TRACE: 07:06:44.901835 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  Zq0zwo dtls TRACE: 07:06:44.901876 handshaker.go:151: [handshake:server] Flight 2: Waiting
  Vvlb7y dtls TRACE: 07:06:44.973972 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:33136
  Zq0zwo dtls TRACE: 07:06:45.109429 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  Zq0zwo dtls TRACE: 07:06:45.109469 handshaker.go:151: [handshake:server] Flight 4: Preparing
  Zq0zwo dtls TRACE: 07:06:45.109481 handshaker.go:151: [handshake:server] Flight 4: Sending
  Zq0zwo dtls TRACE: 07:06:45.109490 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  Zq0zwo dtls TRACE: 07:06:45.109502 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  Zq0zwo dtls TRACE: 07:06:45.109510 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  Zq0zwo dtls TRACE: 07:06:45.109536 handshaker.go:151: [handshake:server] Flight 4: Waiting
  Zq0zwo dtls DEBUG: 07:06:45.324007 conn.go:722: CipherSuite not initialized, queuing packet
  Zq0zwo dtls DEBUG: 07:06:45.324054 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  Zq0zwo dtls TRACE: 07:06:45.324197 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  Zq0zwo dtls TRACE: 07:06:45.324257 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  Zq0zwo dtls TRACE: 07:06:45.324279 handshaker.go:151: [handshake:server] Flight 6: Preparing
  Zq0zwo dtls TRACE: 07:06:45.324322 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  Zq0zwo dtls TRACE: 07:06:45.324342 handshaker.go:151: [handshake:server] Flight 6: Sending
  Zq0zwo dtls TRACE: 07:06:45.324366 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  Zq0zwo dtls TRACE: 07:06:45.324483 handshaker.go:151: [handshake:server] Flight 6: Finished
  Zq0zwo dtls TRACE: 07:06:45.324511 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:33751
  LLLqED dtls TRACE: 07:06:45.324688 handshaker.go:151: [handshake:server] Flight 0: Preparing
  LLLqED dtls TRACE: 07:06:45.324717 handshaker.go:151: [handshake:server] Flight 0: Sending
  LLLqED dtls TRACE: 07:06:45.324731 handshaker.go:151: [handshake:server] Flight 0: Waiting
  LLLqED dtls TRACE: 07:06:45.324963 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  LLLqED dtls TRACE: 07:06:45.324979 handshaker.go:151: [handshake:server] Flight 2: Preparing
  LLLqED dtls TRACE: 07:06:45.324993 handshaker.go:151: [handshake:server] Flight 2: Sending
  LLLqED dtls TRACE: 07:06:45.325010 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  LLLqED dtls TRACE: 07:06:45.325094 handshaker.go:151: [handshake:server] Flight 2: Waiting
  LLLqED dtls TRACE: 07:06:45.325142 handshaker.go:274: [handshake:server] Flight 2 -> Flight 2
  LLLqED dtls TRACE: 07:06:45.325158 handshaker.go:151: [handshake:server] Flight 2: Preparing
  LLLqED dtls TRACE: 07:06:45.325171 handshaker.go:151: [handshake:server] Flight 2: Sending
  LLLqED dtls TRACE: 07:06:45.325187 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  LLLqED dtls TRACE: 07:06:45.325237 handshaker.go:151: [handshake:server] Flight 2: Waiting
  Zq0zwo dtls TRACE: 07:06:45.404325 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:33751
  LLLqED dtls TRACE: 07:06:45.469122 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  LLLqED dtls TRACE: 07:06:45.469136 handshaker.go:151: [handshake:server] Flight 4: Preparing
  LLLqED dtls TRACE: 07:06:45.469142 handshaker.go:151: [handshake:server] Flight 4: Sending
  LLLqED dtls TRACE: 07:06:45.469148 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  LLLqED dtls TRACE: 07:06:45.469155 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  LLLqED dtls TRACE: 07:06:45.469159 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  LLLqED dtls TRACE: 07:06:45.469179 handshaker.go:151: [handshake:server] Flight 4: Waiting
  LLLqED dtls DEBUG: 07:06:45.525890 conn.go:722: CipherSuite not initialized, queuing packet
  LLLqED dtls DEBUG: 07:06:45.526599 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  LLLqED dtls TRACE: 07:06:45.526670 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  LLLqED dtls TRACE: 07:06:45.526687 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  LLLqED dtls TRACE: 07:06:45.526699 handshaker.go:151: [handshake:server] Flight 6: Preparing
  LLLqED dtls TRACE: 07:06:45.526717 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  LLLqED dtls TRACE: 07:06:45.526725 handshaker.go:151: [handshake:server] Flight 6: Sending
  LLLqED dtls TRACE: 07:06:45.526735 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  LLLqED dtls TRACE: 07:06:45.526776 handshaker.go:151: [handshake:server] Flight 6: Finished
  LLLqED dtls TRACE: 07:06:45.526969 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:51096
  N6wjK9 dtls TRACE: 07:06:45.527045 handshaker.go:151: [handshake:server] Flight 0: Preparing
  N6wjK9 dtls TRACE: 07:06:45.527057 handshaker.go:151: [handshake:server] Flight 0: Sending
  N6wjK9 dtls TRACE: 07:06:45.527064 handshaker.go:151: [handshake:server] Flight 0: Waiting
  N6wjK9 dtls TRACE: 07:06:45.527111 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  N6wjK9 dtls TRACE: 07:06:45.527118 handshaker.go:151: [handshake:server] Flight 2: Preparing
  N6wjK9 dtls TRACE: 07:06:45.527261 handshaker.go:151: [handshake:server] Flight 2: Sending
  N6wjK9 dtls TRACE: 07:06:45.527272 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  N6wjK9 dtls TRACE: 07:06:45.527312 handshaker.go:151: [handshake:server] Flight 2: Waiting
  N6wjK9 dtls TRACE: 07:06:45.527388 handshaker.go:274: [handshake:server] Flight 2 -> Flight 2
  N6wjK9 dtls TRACE: 07:06:45.527395 handshaker.go:151: [handshake:server] Flight 2: Preparing
  N6wjK9 dtls TRACE: 07:06:45.527405 handshaker.go:151: [handshake:server] Flight 2: Sending
  N6wjK9 dtls TRACE: 07:06:45.527412 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  N6wjK9 dtls TRACE: 07:06:45.527430 handshaker.go:151: [handshake:server] Flight 2: Waiting
  N6wjK9 dtls TRACE: 07:06:45.717704 handshaker.go:274: [handshake:server] Flight 2 -> Flight 2
  N6wjK9 dtls TRACE: 07:06:45.717746 handshaker.go:151: [handshake:server] Flight 2: Preparing
  N6wjK9 dtls TRACE: 07:06:45.717857 handshaker.go:151: [handshake:server] Flight 2: Sending
  N6wjK9 dtls TRACE: 07:06:45.717885 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  N6wjK9 dtls TRACE: 07:06:45.718003 handshaker.go:151: [handshake:server] Flight 2: Waiting
  LLLqED dtls TRACE: 07:06:45.759358 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:51096
  N6wjK9 dtls TRACE: 07:06:45.945478 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  N6wjK9 dtls TRACE: 07:06:45.945522 handshaker.go:151: [handshake:server] Flight 4: Preparing
  N6wjK9 dtls TRACE: 07:06:45.945549 handshaker.go:151: [handshake:server] Flight 4: Sending
  N6wjK9 dtls TRACE: 07:06:45.945771 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  N6wjK9 dtls TRACE: 07:06:45.945838 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  N6wjK9 dtls TRACE: 07:06:45.945876 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  N6wjK9 dtls TRACE: 07:06:45.945999 handshaker.go:151: [handshake:server] Flight 4: Waiting
  N6wjK9 dtls DEBUG: 07:06:46.091657 conn.go:722: CipherSuite not initialized, queuing packet
  N6wjK9 dtls DEBUG: 07:06:46.091926 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  N6wjK9 dtls TRACE: 07:06:46.092251 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  N6wjK9 dtls TRACE: 07:06:46.092477 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  N6wjK9 dtls TRACE: 07:06:46.092593 handshaker.go:151: [handshake:server] Flight 6: Preparing
  N6wjK9 dtls TRACE: 07:06:46.092664 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  N6wjK9 dtls TRACE: 07:06:46.092705 handshaker.go:151: [handshake:server] Flight 6: Sending
  N6wjK9 dtls TRACE: 07:06:46.092742 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  N6wjK9 dtls TRACE: 07:06:46.092915 handshaker.go:151: [handshake:server] Flight 6: Finished
  N6wjK9 dtls TRACE: 07:06:46.093143 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:33292
  RkSnKM dtls TRACE: 07:06:46.093480 handshaker.go:151: [handshake:server] Flight 0: Preparing
  RkSnKM dtls TRACE: 07:06:46.093568 handshaker.go:151: [handshake:server] Flight 0: Sending
  RkSnKM dtls TRACE: 07:06:46.093628 handshaker.go:151: [handshake:server] Flight 0: Waiting
  RkSnKM dtls TRACE: 07:06:46.093885 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  RkSnKM dtls TRACE: 07:06:46.093978 handshaker.go:151: [handshake:server] Flight 2: Preparing
  RkSnKM dtls TRACE: 07:06:46.094011 handshaker.go:151: [handshake:server] Flight 2: Sending
  RkSnKM dtls TRACE: 07:06:46.094060 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  RkSnKM dtls TRACE: 07:06:46.094190 handshaker.go:151: [handshake:server] Flight 2: Waiting
  RkSnKM dtls TRACE: 07:06:46.094370 handshaker.go:274: [handshake:server] Flight 2 -> Flight 2
  RkSnKM dtls TRACE: 07:06:46.094521 handshaker.go:151: [handshake:server] Flight 2: Preparing
  RkSnKM dtls TRACE: 07:06:46.094557 handshaker.go:151: [handshake:server] Flight 2: Sending
  RkSnKM dtls TRACE: 07:06:46.094604 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  RkSnKM dtls TRACE: 07:06:46.094696 handshaker.go:151: [handshake:server] Flight 2: Waiting
  RkSnKM dtls TRACE: 07:06:46.095101 handshaker.go:274: [handshake:server] Flight 2 -> Flight 2
  RkSnKM dtls TRACE: 07:06:46.095293 handshaker.go:151: [handshake:server] Flight 2: Preparing
  RkSnKM dtls TRACE: 07:06:46.095357 handshaker.go:151: [handshake:server] Flight 2: Sending
  RkSnKM dtls TRACE: 07:06:46.095392 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  RkSnKM dtls TRACE: 07:06:46.095547 handshaker.go:151: [handshake:server] Flight 2: Waiting
  N6wjK9 dtls TRACE: 07:06:46.194144 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:33292
  RkSnKM dtls TRACE: 07:06:46.197221 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  RkSnKM dtls TRACE: 07:06:46.197344 handshaker.go:151: [handshake:server] Flight 4: Preparing
  RkSnKM dtls TRACE: 07:06:46.197377 handshaker.go:151: [handshake:server] Flight 4: Sending
  RkSnKM dtls TRACE: 07:06:46.197417 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  RkSnKM dtls TRACE: 07:06:46.197447 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  RkSnKM dtls TRACE: 07:06:46.197471 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  RkSnKM dtls TRACE: 07:06:46.197559 handshaker.go:151: [handshake:server] Flight 4: Waiting
  RkSnKM dtls DEBUG: 07:06:46.435688 conn.go:722: CipherSuite not initialized, queuing packet
  RkSnKM dtls DEBUG: 07:06:46.435864 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  RkSnKM dtls TRACE: 07:06:46.436062 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  RkSnKM dtls TRACE: 07:06:46.436132 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  RkSnKM dtls TRACE: 07:06:46.436193 handshaker.go:151: [handshake:server] Flight 6: Preparing
  RkSnKM dtls TRACE: 07:06:46.436253 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  RkSnKM dtls TRACE: 07:06:46.436276 handshaker.go:151: [handshake:server] Flight 6: Sending
  RkSnKM dtls TRACE: 07:06:46.436301 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  RkSnKM dtls TRACE: 07:06:46.436410 handshaker.go:151: [handshake:server] Flight 6: Finished
  RkSnKM dtls TRACE: 07:06:46.436557 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:56420
  RSDyZy dtls TRACE: 07:06:46.436758 handshaker.go:151: [handshake:server] Flight 0: Preparing
  RSDyZy dtls TRACE: 07:06:46.436786 handshaker.go:151: [handshake:server] Flight 0: Sending
  RSDyZy dtls TRACE: 07:06:46.436800 handshaker.go:151: [handshake:server] Flight 0: Waiting
  RSDyZy dtls TRACE: 07:06:46.436945 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  RSDyZy dtls TRACE: 07:06:46.436960 handshaker.go:151: [handshake:server] Flight 2: Preparing
  RSDyZy dtls TRACE: 07:06:46.436973 handshaker.go:151: [handshake:server] Flight 2: Sending
  RSDyZy dtls TRACE: 07:06:46.436989 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  RSDyZy dtls TRACE: 07:06:46.437053 handshaker.go:151: [handshake:server] Flight 2: Waiting
  RSDyZy dtls TRACE: 07:06:46.437103 handshaker.go:274: [handshake:server] Flight 2 -> Flight 2
  RSDyZy dtls TRACE: 07:06:46.437147 handshaker.go:151: [handshake:server] Flight 2: Preparing
  RSDyZy dtls TRACE: 07:06:46.437171 handshaker.go:151: [handshake:server] Flight 2: Sending
  RSDyZy dtls TRACE: 07:06:46.437194 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  RSDyZy dtls TRACE: 07:06:46.437278 handshaker.go:151: [handshake:server] Flight 2: Waiting
  RSDyZy dtls TRACE: 07:06:46.437360 handshaker.go:274: [handshake:server] Flight 2 -> Flight 2
  RSDyZy dtls TRACE: 07:06:46.437398 handshaker.go:151: [handshake:server] Flight 2: Preparing
  RSDyZy dtls TRACE: 07:06:46.437425 handshaker.go:151: [handshake:server] Flight 2: Sending
  RSDyZy dtls TRACE: 07:06:46.437446 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  RSDyZy dtls TRACE: 07:06:46.437514 handshaker.go:151: [handshake:server] Flight 2: Waiting
  RkSnKM dtls TRACE: 07:06:46.561956 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:56420
  RSDyZy dtls TRACE: 07:06:46.661991 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  RSDyZy dtls TRACE: 07:06:46.662007 handshaker.go:151: [handshake:server] Flight 4: Preparing
  RSDyZy dtls TRACE: 07:06:46.662013 handshaker.go:151: [handshake:server] Flight 4: Sending
  RSDyZy dtls TRACE: 07:06:46.662019 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  RSDyZy dtls TRACE: 07:06:46.662024 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  RSDyZy dtls TRACE: 07:06:46.662028 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  RSDyZy dtls TRACE: 07:06:46.662052 handshaker.go:151: [handshake:server] Flight 4: Waiting
  RSDyZy dtls DEBUG: 07:06:46.800752 conn.go:722: CipherSuite not initialized, queuing packet
  RSDyZy dtls DEBUG: 07:06:46.800779 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  RSDyZy dtls TRACE: 07:06:46.800880 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  RSDyZy dtls TRACE: 07:06:46.800920 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  RSDyZy dtls TRACE: 07:06:46.800935 handshaker.go:151: [handshake:server] Flight 6: Preparing
  RSDyZy dtls TRACE: 07:06:46.800961 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  RSDyZy dtls TRACE: 07:06:46.800972 handshaker.go:151: [handshake:server] Flight 6: Sending
  RSDyZy dtls TRACE: 07:06:46.800986 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  RSDyZy dtls TRACE: 07:06:46.801067 handshaker.go:151: [handshake:server] Flight 6: Finished
  RSDyZy dtls TRACE: 07:06:46.801091 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:43784
  RSDyZy dtls TRACE: 07:06:46.859031 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:43784
  ```

  </details>

  <details><summary>Client-side log with 1 concurrent listener (original situation).</summary>

  ```shell
  $ mkdir -p build && go build -o build/ ./examples/dial/psk && for i in {1..10}; do ./build/psk <<< 'exit' &; done
  [2] 14344
  [3] 14345
  [4] 14346
  [5] 14350
  Ee9RNy dtls TRACE: 07:06:43.403699 handshaker.go:151: [handshake:client] Flight 1: Preparing
  kns6kD dtls TRACE: 07:06:43.403699 handshaker.go:151: [handshake:client] Flight 1: Preparing
  kns6kD dtls TRACE: 07:06:43.403747 handshaker.go:151: [handshake:client] Flight 1: Sending
  kns6kD dtls TRACE: 07:06:43.403765 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  kns6kD dtls TRACE: 07:06:43.403809 handshaker.go:151: [handshake:client] Flight 1: Waiting
  [6] 14358
  Ee9RNy dtls TRACE: 07:06:43.403737 handshaker.go:151: [handshake:client] Flight 1: Sending
  Ee9RNy dtls TRACE: 07:06:43.403931 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  Ee9RNy dtls TRACE: 07:06:43.403971 handshaker.go:151: [handshake:client] Flight 1: Waiting
  kns6kD dtls TRACE: 07:06:43.404171 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  kns6kD dtls TRACE: 07:06:43.404186 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 230ms
  [7] 14364
  hKskmM dtls TRACE: 07:06:43.405227 handshaker.go:151: [handshake:client] Flight 1: Preparing
  hKskmM dtls TRACE: 07:06:43.405291 handshaker.go:151: [handshake:client] Flight 1: Sending
  hKskmM dtls TRACE: 07:06:43.405313 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  hKskmM dtls TRACE: 07:06:43.405468 handshaker.go:151: [handshake:client] Flight 1: Waiting
  18WgqG dtls TRACE: 07:06:43.405693 handshaker.go:151: [handshake:client] Flight 1: Preparing
  18WgqG dtls TRACE: 07:06:43.405739 handshaker.go:151: [handshake:client] Flight 1: Sending
  18WgqG dtls TRACE: 07:06:43.405753 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  18WgqG dtls TRACE: 07:06:43.405789 handshaker.go:151: [handshake:client] Flight 1: Waiting
  hOMSjq dtls TRACE: 07:06:43.406252 handshaker.go:151: [handshake:client] Flight 1: Preparing
  hOMSjq dtls TRACE: 07:06:43.406356 handshaker.go:151: [handshake:client] Flight 1: Sending
  hOMSjq dtls TRACE: 07:06:43.406372 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  hOMSjq dtls TRACE: 07:06:43.406415 handshaker.go:151: [handshake:client] Flight 1: Waiting
  [8] 14367
  3dHGB0 dtls TRACE: 07:06:43.407637 handshaker.go:151: [handshake:client] Flight 1: Preparing
  [9] 14379
  kRUTE5 dtls TRACE: 07:06:43.408841 handshaker.go:151: [handshake:client] Flight 1: Preparing
  kRUTE5 dtls TRACE: 07:06:43.410522 handshaker.go:151: [handshake:client] Flight 1: Sending
  kRUTE5 dtls TRACE: 07:06:43.410788 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  kRUTE5 dtls TRACE: 07:06:43.410845 handshaker.go:151: [handshake:client] Flight 1: Waiting
  Dn6shc dtls TRACE: 07:06:43.409871 handshaker.go:151: [handshake:client] Flight 1: Preparing
  Dn6shc dtls TRACE: 07:06:43.410927 handshaker.go:151: [handshake:client] Flight 1: Sending
  Dn6shc dtls TRACE: 07:06:43.410944 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  Dn6shc dtls TRACE: 07:06:43.411043 handshaker.go:151: [handshake:client] Flight 1: Waiting
  3dHGB0 dtls TRACE: 07:06:43.411340 handshaker.go:151: [handshake:client] Flight 1: Sending
  3dHGB0 dtls TRACE: 07:06:43.411359 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  3dHGB0 dtls TRACE: 07:06:43.411467 handshaker.go:151: [handshake:client] Flight 1: Waiting
  [10] 14389
  [11] 14391
  v0YnEL dtls TRACE: 07:06:43.421379 handshaker.go:151: [handshake:client] Flight 1: Preparing
  v0YnEL dtls TRACE: 07:06:43.421439 handshaker.go:151: [handshake:client] Flight 1: Sending
  v0YnEL dtls TRACE: 07:06:43.421455 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  v0YnEL dtls TRACE: 07:06:43.421499 handshaker.go:151: [handshake:client] Flight 1: Waiting
  CRJpcH dtls TRACE: 07:06:43.432648 handshaker.go:151: [handshake:client] Flight 1: Preparing
  CRJpcH dtls TRACE: 07:06:43.432696 handshaker.go:151: [handshake:client] Flight 1: Sending
  CRJpcH dtls TRACE: 07:06:43.432708 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  CRJpcH dtls TRACE: 07:06:43.432765 handshaker.go:151: [handshake:client] Flight 1: Waiting
  kns6kD dtls TRACE: 07:06:43.634462 handshaker.go:151: [handshake:client] Flight 3: Preparing
  kns6kD dtls TRACE: 07:06:43.634544 handshaker.go:151: [handshake:client] Flight 3: Sending
  kns6kD dtls TRACE: 07:06:43.634643 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  kns6kD dtls TRACE: 07:06:43.634813 handshaker.go:151: [handshake:client] Flight 3: Waiting
  kns6kD dtls TRACE: 07:06:43.635733 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  kns6kD dtls TRACE: 07:06:43.635966 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  kns6kD dtls TRACE: 07:06:43.636031 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 142ms
  kns6kD dtls TRACE: 07:06:43.778480 handshaker.go:151: [handshake:client] Flight 5: Preparing
  kns6kD dtls TRACE: 07:06:43.778757 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  kns6kD dtls TRACE: 07:06:43.778785 handshaker.go:151: [handshake:client] Flight 5: Sending
  kns6kD dtls TRACE: 07:06:43.778841 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  kns6kD dtls TRACE: 07:06:43.778874 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  kns6kD dtls TRACE: 07:06:43.779072 handshaker.go:151: [handshake:client] Flight 5: Waiting
  kns6kD dtls TRACE: 07:06:43.780749 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  kns6kD dtls TRACE: 07:06:43.780865 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  kns6kD dtls TRACE: 07:06:43.780896 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 148ms
  Ee9RNy dtls TRACE: 07:06:43.781181 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  Ee9RNy dtls TRACE: 07:06:43.781478 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 133ms
  Ee9RNy dtls TRACE: 07:06:43.914694 handshaker.go:151: [handshake:client] Flight 3: Preparing
  Ee9RNy dtls TRACE: 07:06:43.914818 handshaker.go:151: [handshake:client] Flight 3: Sending
  Ee9RNy dtls TRACE: 07:06:43.914921 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  Ee9RNy dtls TRACE: 07:06:43.915196 handshaker.go:151: [handshake:client] Flight 3: Waiting
  Ee9RNy dtls TRACE: 07:06:43.915739 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  Ee9RNy dtls TRACE: 07:06:43.915786 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  Ee9RNy dtls TRACE: 07:06:43.915852 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 71ms
  kns6kD dtls TRACE: 07:06:43.929110 handshaker.go:151: [handshake:client] Flight 5: Finished
  kns6kD dtls TRACE: 07:06:43.929195 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF

  [4]    14346 done       ./build/psk <<< 'exit'
  Ee9RNy dtls TRACE: 07:06:43.989005 handshaker.go:151: [handshake:client] Flight 5: Preparing
  Ee9RNy dtls TRACE: 07:06:43.989337 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  Ee9RNy dtls TRACE: 07:06:43.989381 handshaker.go:151: [handshake:client] Flight 5: Sending
  Ee9RNy dtls TRACE: 07:06:43.989417 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  Ee9RNy dtls TRACE: 07:06:43.989461 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  Ee9RNy dtls TRACE: 07:06:43.989681 handshaker.go:151: [handshake:client] Flight 5: Waiting
  Ee9RNy dtls TRACE: 07:06:43.993033 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  hKskmM dtls TRACE: 07:06:43.994667 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  hKskmM dtls TRACE: 07:06:43.995828 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 85ms
  Ee9RNy dtls TRACE: 07:06:43.995765 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  Ee9RNy dtls TRACE: 07:06:43.995971 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 197ms
  hKskmM dtls TRACE: 07:06:44.081011 handshaker.go:151: [handshake:client] Flight 3: Preparing
  hKskmM dtls TRACE: 07:06:44.081037 handshaker.go:151: [handshake:client] Flight 3: Sending
  hKskmM dtls TRACE: 07:06:44.081063 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  hKskmM dtls TRACE: 07:06:44.081114 handshaker.go:151: [handshake:client] Flight 3: Waiting
  hKskmM dtls TRACE: 07:06:44.081509 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  hKskmM dtls TRACE: 07:06:44.081545 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  hKskmM dtls TRACE: 07:06:44.081556 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 156ms
  Ee9RNy dtls TRACE: 07:06:44.193023 handshaker.go:151: [handshake:client] Flight 5: Finished
  Ee9RNy dtls TRACE: 07:06:44.193048 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF
  panic: EOF

  goroutine 35 [running]:
  github.com/pion/dtls/v2/examples/util.Check(0x652d00, 0xc000012120)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:65 +0x1b4
  github.com/pion/dtls/v2/examples/util.Chat.func1(0x654180, 0xc00009e580)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:31 +0x152
  created by github.com/pion/dtls/v2/examples/util.Chat
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:26 +0x64

  [3]    14345 exit 2     ./build/psk <<< 'exit'
  hKskmM dtls TRACE: 07:06:44.237670 handshaker.go:151: [handshake:client] Flight 5: Preparing
  hKskmM dtls TRACE: 07:06:44.237726 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  hKskmM dtls TRACE: 07:06:44.237738 handshaker.go:151: [handshake:client] Flight 5: Sending
  hKskmM dtls TRACE: 07:06:44.237747 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  hKskmM dtls TRACE: 07:06:44.237758 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  hKskmM dtls TRACE: 07:06:44.237820 handshaker.go:151: [handshake:client] Flight 5: Waiting
  hKskmM dtls TRACE: 07:06:44.238045 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  hKskmM dtls TRACE: 07:06:44.238082 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  hKskmM dtls TRACE: 07:06:44.238112 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 197ms
  18WgqG dtls TRACE: 07:06:44.238283 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  18WgqG dtls TRACE: 07:06:44.238307 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 172ms
  hOMSjq dtls TRACE: 07:06:44.406714 handshaker.go:165: [handshake:client] Flight 1: Sending but sleeping for 126ms
  kRUTE5 dtls TRACE: 07:06:44.411145 handshaker.go:165: [handshake:client] Flight 1: Sending but sleeping for 59ms
  18WgqG dtls TRACE: 07:06:44.411290 handshaker.go:151: [handshake:client] Flight 3: Preparing
  18WgqG dtls TRACE: 07:06:44.411323 handshaker.go:151: [handshake:client] Flight 3: Sending
  Dn6shc dtls TRACE: 07:06:44.411381 handshaker.go:165: [handshake:client] Flight 1: Sending but sleeping for 129ms
  18WgqG dtls TRACE: 07:06:44.411554 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  18WgqG dtls TRACE: 07:06:44.411923 handshaker.go:151: [handshake:client] Flight 3: Waiting
  3dHGB0 dtls TRACE: 07:06:44.411727 handshaker.go:165: [handshake:client] Flight 1: Sending but sleeping for 74ms
  18WgqG dtls TRACE: 07:06:44.412780 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  18WgqG dtls TRACE: 07:06:44.412880 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  18WgqG dtls TRACE: 07:06:44.413675 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 156ms
  v0YnEL dtls TRACE: 07:06:44.422208 handshaker.go:165: [handshake:client] Flight 1: Sending but sleeping for 229ms
  CRJpcH dtls TRACE: 07:06:44.433114 handshaker.go:165: [handshake:client] Flight 1: Sending but sleeping for 63ms
  hKskmM dtls TRACE: 07:06:44.435171 handshaker.go:151: [handshake:client] Flight 5: Finished
  hKskmM dtls TRACE: 07:06:44.435282 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF

  [5]    14350 done       ./build/psk <<< 'exit'
  kRUTE5 dtls TRACE: 07:06:44.470443 handshaker.go:151: [handshake:client] Flight 1: Sending
  kRUTE5 dtls TRACE: 07:06:44.470626 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  kRUTE5 dtls TRACE: 07:06:44.470824 handshaker.go:151: [handshake:client] Flight 1: Waiting
  3dHGB0 dtls TRACE: 07:06:44.486268 handshaker.go:151: [handshake:client] Flight 1: Sending
  3dHGB0 dtls TRACE: 07:06:44.486458 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  3dHGB0 dtls TRACE: 07:06:44.486843 handshaker.go:151: [handshake:client] Flight 1: Waiting
  CRJpcH dtls TRACE: 07:06:44.496522 handshaker.go:151: [handshake:client] Flight 1: Sending
  CRJpcH dtls TRACE: 07:06:44.496696 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  CRJpcH dtls TRACE: 07:06:44.496876 handshaker.go:151: [handshake:client] Flight 1: Waiting
  hOMSjq dtls TRACE: 07:06:44.533361 handshaker.go:151: [handshake:client] Flight 1: Sending
  hOMSjq dtls TRACE: 07:06:44.533395 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  hOMSjq dtls TRACE: 07:06:44.533450 handshaker.go:151: [handshake:client] Flight 1: Waiting
  Dn6shc dtls TRACE: 07:06:44.540713 handshaker.go:151: [handshake:client] Flight 1: Sending
  Dn6shc dtls TRACE: 07:06:44.540761 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  Dn6shc dtls TRACE: 07:06:44.540837 handshaker.go:151: [handshake:client] Flight 1: Waiting
  18WgqG dtls TRACE: 07:06:44.569842 handshaker.go:151: [handshake:client] Flight 5: Preparing
  18WgqG dtls TRACE: 07:06:44.569921 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  18WgqG dtls TRACE: 07:06:44.569929 handshaker.go:151: [handshake:client] Flight 5: Sending
  18WgqG dtls TRACE: 07:06:44.569935 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  18WgqG dtls TRACE: 07:06:44.569945 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  18WgqG dtls TRACE: 07:06:44.570024 handshaker.go:151: [handshake:client] Flight 5: Waiting
  18WgqG dtls TRACE: 07:06:44.570406 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  18WgqG dtls TRACE: 07:06:44.570455 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  18WgqG dtls TRACE: 07:06:44.570606 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 182ms
  hOMSjq dtls TRACE: 07:06:44.570876 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  hOMSjq dtls TRACE: 07:06:44.570885 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 101ms
  v0YnEL dtls TRACE: 07:06:44.651473 handshaker.go:151: [handshake:client] Flight 1: Sending
  v0YnEL dtls TRACE: 07:06:44.651497 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  v0YnEL dtls TRACE: 07:06:44.651556 handshaker.go:151: [handshake:client] Flight 1: Waiting
  hOMSjq dtls TRACE: 07:06:44.672042 handshaker.go:151: [handshake:client] Flight 3: Preparing
  hOMSjq dtls TRACE: 07:06:44.672082 handshaker.go:151: [handshake:client] Flight 3: Sending
  hOMSjq dtls TRACE: 07:06:44.672114 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  hOMSjq dtls TRACE: 07:06:44.672181 handshaker.go:151: [handshake:client] Flight 3: Waiting
  hOMSjq dtls TRACE: 07:06:44.672596 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  hOMSjq dtls TRACE: 07:06:44.672786 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  hOMSjq dtls TRACE: 07:06:44.672797 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 226ms
  18WgqG dtls TRACE: 07:06:44.752835 handshaker.go:151: [handshake:client] Flight 5: Finished
  18WgqG dtls TRACE: 07:06:44.752881 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF
  panic: EOF

  goroutine 5 [running]:
  github.com/pion/dtls/v2/examples/util.Check(0x652d00, 0xc000098100)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:65 +0x1b4
  github.com/pion/dtls/v2/examples/util.Chat.func1(0x654180, 0xc0000da580)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:31 +0x152
  created by github.com/pion/dtls/v2/examples/util.Chat
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:26 +0x64

  [2]    14344 exit 2     ./build/psk <<< 'exit'
  hOMSjq dtls TRACE: 07:06:44.898920 handshaker.go:151: [handshake:client] Flight 5: Preparing
  hOMSjq dtls TRACE: 07:06:44.899113 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  hOMSjq dtls TRACE: 07:06:44.899134 handshaker.go:151: [handshake:client] Flight 5: Sending
  hOMSjq dtls TRACE: 07:06:44.899152 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  hOMSjq dtls TRACE: 07:06:44.899177 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  hOMSjq dtls TRACE: 07:06:44.899323 handshaker.go:151: [handshake:client] Flight 5: Waiting
  hOMSjq dtls TRACE: 07:06:44.899887 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  hOMSjq dtls TRACE: 07:06:44.900023 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  hOMSjq dtls TRACE: 07:06:44.900062 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 73ms
  kRUTE5 dtls TRACE: 07:06:44.902052 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  kRUTE5 dtls TRACE: 07:06:44.902088 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 207ms
  hOMSjq dtls TRACE: 07:06:44.973449 handshaker.go:151: [handshake:client] Flight 5: Finished
  hOMSjq dtls TRACE: 07:06:44.973478 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF
  panic: EOF

  goroutine 7 [running]:
  github.com/pion/dtls/v2/examples/util.Check(0x652d00, 0xc000112100)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:65 +0x1b4
  github.com/pion/dtls/v2/examples/util.Chat.func1(0x654180, 0xc000152580)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:31 +0x152
  created by github.com/pion/dtls/v2/examples/util.Chat
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:26 +0x64
  [6]    14358 exit 2     ./build/psk <<< 'exit'
  kRUTE5 dtls TRACE: 07:06:45.109260 handshaker.go:151: [handshake:client] Flight 3: Preparing
  kRUTE5 dtls TRACE: 07:06:45.109299 handshaker.go:151: [handshake:client] Flight 3: Sending
  kRUTE5 dtls TRACE: 07:06:45.109324 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  kRUTE5 dtls TRACE: 07:06:45.109757 handshaker.go:151: [handshake:client] Flight 3: Waiting
  kRUTE5 dtls TRACE: 07:06:45.109811 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  kRUTE5 dtls TRACE: 07:06:45.109828 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  kRUTE5 dtls TRACE: 07:06:45.109838 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 213ms
  kRUTE5 dtls TRACE: 07:06:45.323163 handshaker.go:151: [handshake:client] Flight 5: Preparing
  kRUTE5 dtls TRACE: 07:06:45.323512 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  kRUTE5 dtls TRACE: 07:06:45.323538 handshaker.go:151: [handshake:client] Flight 5: Sending
  kRUTE5 dtls TRACE: 07:06:45.323561 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  kRUTE5 dtls TRACE: 07:06:45.323591 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  kRUTE5 dtls TRACE: 07:06:45.323839 handshaker.go:151: [handshake:client] Flight 5: Waiting
  kRUTE5 dtls TRACE: 07:06:45.324543 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  kRUTE5 dtls TRACE: 07:06:45.324653 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  kRUTE5 dtls TRACE: 07:06:45.324678 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 79ms
  Dn6shc dtls TRACE: 07:06:45.325450 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  Dn6shc dtls TRACE: 07:06:45.325487 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 143ms
  kRUTE5 dtls TRACE: 07:06:45.403829 handshaker.go:151: [handshake:client] Flight 5: Finished
  kRUTE5 dtls TRACE: 07:06:45.403906 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF

  [8]    14367 done       ./build/psk <<< 'exit'
  Dn6shc dtls TRACE: 07:06:45.468896 handshaker.go:151: [handshake:client] Flight 3: Preparing
  Dn6shc dtls TRACE: 07:06:45.468940 handshaker.go:151: [handshake:client] Flight 3: Sending
  Dn6shc dtls TRACE: 07:06:45.468975 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  Dn6shc dtls TRACE: 07:06:45.469029 handshaker.go:151: [handshake:client] Flight 3: Waiting
  Dn6shc dtls TRACE: 07:06:45.469245 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  Dn6shc dtls TRACE: 07:06:45.469254 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  Dn6shc dtls TRACE: 07:06:45.469260 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 56ms
  3dHGB0 dtls TRACE: 07:06:45.487020 handshaker.go:165: [handshake:client] Flight 1: Sending but sleeping for 230ms
  CRJpcH dtls TRACE: 07:06:45.497156 handshaker.go:165: [handshake:client] Flight 1: Sending but sleeping for 115ms
  Dn6shc dtls TRACE: 07:06:45.525712 handshaker.go:151: [handshake:client] Flight 5: Preparing
  Dn6shc dtls TRACE: 07:06:45.525768 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  Dn6shc dtls TRACE: 07:06:45.525773 handshaker.go:151: [handshake:client] Flight 5: Sending
  Dn6shc dtls TRACE: 07:06:45.525778 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  Dn6shc dtls TRACE: 07:06:45.525785 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  Dn6shc dtls TRACE: 07:06:45.525829 handshaker.go:151: [handshake:client] Flight 5: Waiting
  Dn6shc dtls TRACE: 07:06:45.527468 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  Dn6shc dtls TRACE: 07:06:45.527726 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  Dn6shc dtls TRACE: 07:06:45.527774 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 230ms
  CRJpcH dtls TRACE: 07:06:45.612452 handshaker.go:151: [handshake:client] Flight 1: Sending
  CRJpcH dtls TRACE: 07:06:45.612566 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  CRJpcH dtls TRACE: 07:06:45.612814 handshaker.go:151: [handshake:client] Flight 1: Waiting
  v0YnEL dtls TRACE: 07:06:45.651930 handshaker.go:165: [handshake:client] Flight 1: Sending but sleeping for 186ms
  3dHGB0 dtls TRACE: 07:06:45.717227 handshaker.go:151: [handshake:client] Flight 1: Sending
  3dHGB0 dtls TRACE: 07:06:45.717374 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  3dHGB0 dtls TRACE: 07:06:45.717553 handshaker.go:151: [handshake:client] Flight 1: Waiting
  3dHGB0 dtls TRACE: 07:06:45.717636 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  3dHGB0 dtls TRACE: 07:06:45.717665 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 227ms
  Dn6shc dtls TRACE: 07:06:45.758005 handshaker.go:151: [handshake:client] Flight 5: Finished
  Dn6shc dtls TRACE: 07:06:45.758331 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully

  [9]    14379 done       ./build/psk <<< 'exit'
  v0YnEL dtls TRACE: 07:06:45.838075 handshaker.go:151: [handshake:client] Flight 1: Sending
  v0YnEL dtls TRACE: 07:06:45.838122 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  v0YnEL dtls TRACE: 07:06:45.838179 handshaker.go:151: [handshake:client] Flight 1: Waiting
  3dHGB0 dtls TRACE: 07:06:45.944902 handshaker.go:151: [handshake:client] Flight 3: Preparing
  3dHGB0 dtls TRACE: 07:06:45.945060 handshaker.go:151: [handshake:client] Flight 3: Sending
  3dHGB0 dtls TRACE: 07:06:45.945145 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  3dHGB0 dtls TRACE: 07:06:45.945289 handshaker.go:151: [handshake:client] Flight 3: Waiting
  3dHGB0 dtls TRACE: 07:06:45.950607 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  3dHGB0 dtls TRACE: 07:06:45.950661 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  3dHGB0 dtls TRACE: 07:06:45.950685 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 140ms
  3dHGB0 dtls TRACE: 07:06:46.090991 handshaker.go:151: [handshake:client] Flight 5: Preparing
  3dHGB0 dtls TRACE: 07:06:46.091192 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  3dHGB0 dtls TRACE: 07:06:46.091245 handshaker.go:151: [handshake:client] Flight 5: Sending
  3dHGB0 dtls TRACE: 07:06:46.091271 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  3dHGB0 dtls TRACE: 07:06:46.091303 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  3dHGB0 dtls TRACE: 07:06:46.091485 handshaker.go:151: [handshake:client] Flight 5: Waiting
  3dHGB0 dtls TRACE: 07:06:46.093039 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  3dHGB0 dtls TRACE: 07:06:46.095788 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  v0YnEL dtls TRACE: 07:06:46.094959 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  v0YnEL dtls TRACE: 07:06:46.096051 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 100ms
  3dHGB0 dtls TRACE: 07:06:46.095969 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 97ms
  3dHGB0 dtls TRACE: 07:06:46.193507 handshaker.go:151: [handshake:client] Flight 5: Finished
  3dHGB0 dtls TRACE: 07:06:46.193634 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF
  v0YnEL dtls TRACE: 07:06:46.196336 handshaker.go:151: [handshake:client] Flight 3: Preparing
  v0YnEL dtls TRACE: 07:06:46.196456 handshaker.go:151: [handshake:client] Flight 3: Sending
  v0YnEL dtls TRACE: 07:06:46.196549 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  v0YnEL dtls TRACE: 07:06:46.196832 handshaker.go:151: [handshake:client] Flight 3: Waiting

  [7]    14364 done       ./build/psk <<< 'exit'
  v0YnEL dtls TRACE: 07:06:46.197972 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  v0YnEL dtls TRACE: 07:06:46.198026 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  v0YnEL dtls TRACE: 07:06:46.198056 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 237ms
  v0YnEL dtls TRACE: 07:06:46.435218 handshaker.go:151: [handshake:client] Flight 5: Preparing
  v0YnEL dtls TRACE: 07:06:46.435399 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  v0YnEL dtls TRACE: 07:06:46.435419 handshaker.go:151: [handshake:client] Flight 5: Sending
  v0YnEL dtls TRACE: 07:06:46.435434 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  v0YnEL dtls TRACE: 07:06:46.435455 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  v0YnEL dtls TRACE: 07:06:46.435578 handshaker.go:151: [handshake:client] Flight 5: Waiting
  v0YnEL dtls TRACE: 07:06:46.436694 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  v0YnEL dtls TRACE: 07:06:46.436886 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  v0YnEL dtls TRACE: 07:06:46.436908 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 124ms
  CRJpcH dtls TRACE: 07:06:46.437725 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  CRJpcH dtls TRACE: 07:06:46.437750 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 224ms
  v0YnEL dtls TRACE: 07:06:46.561713 handshaker.go:151: [handshake:client] Flight 5: Finished
  v0YnEL dtls TRACE: 07:06:46.561741 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF
  panic: EOF

  goroutine 20 [running]:
  github.com/pion/dtls/v2/examples/util.Check(0x652d00, 0xc000012120)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:65 +0x1b4
  github.com/pion/dtls/v2/examples/util.Chat.func1(0x654180, 0xc00011e580)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:31 +0x152
  created by github.com/pion/dtls/v2/examples/util.Chat
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:26 +0x64
  [11]  + 14391 exit 2     ./build/psk <<< 'exit'
  CRJpcH dtls TRACE: 07:06:46.661852 handshaker.go:151: [handshake:client] Flight 3: Preparing
  CRJpcH dtls TRACE: 07:06:46.661869 handshaker.go:151: [handshake:client] Flight 3: Sending
  CRJpcH dtls TRACE: 07:06:46.661896 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  CRJpcH dtls TRACE: 07:06:46.661937 handshaker.go:151: [handshake:client] Flight 3: Waiting
  CRJpcH dtls TRACE: 07:06:46.662110 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  CRJpcH dtls TRACE: 07:06:46.662138 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  CRJpcH dtls TRACE: 07:06:46.662154 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 138ms
  CRJpcH dtls TRACE: 07:06:46.800367 handshaker.go:151: [handshake:client] Flight 5: Preparing
  CRJpcH dtls TRACE: 07:06:46.800504 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  CRJpcH dtls TRACE: 07:06:46.800518 handshaker.go:151: [handshake:client] Flight 5: Sending
  CRJpcH dtls TRACE: 07:06:46.800531 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  CRJpcH dtls TRACE: 07:06:46.800548 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  CRJpcH dtls TRACE: 07:06:46.800673 handshaker.go:151: [handshake:client] Flight 5: Waiting
  CRJpcH dtls TRACE: 07:06:46.801125 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  CRJpcH dtls TRACE: 07:06:46.801194 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  CRJpcH dtls TRACE: 07:06:46.801211 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 57ms
  CRJpcH dtls TRACE: 07:06:46.858458 handshaker.go:151: [handshake:client] Flight 5: Finished
  CRJpcH dtls TRACE: 07:06:46.858525 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF
  panic: EOF

  goroutine 13 [running]:
  github.com/pion/dtls/v2/examples/util.Check(0x652d00, 0xc000012120)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:65 +0x1b4
  github.com/pion/dtls/v2/examples/util.Chat.func1(0x654180, 0xc00009e580)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:31 +0x152
  created by github.com/pion/dtls/v2/examples/util.Chat
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:26 +0x64

  [10]  + 14389 exit 2     ./build/psk <<< 'exit'
  ```

  </details>

- With patch applied.

  <details><summary>Server-side log with 5 concurrent listeners.</summary>

  ```shell
  $ go run ./examples/listen/psk/
  Listening
  WhbypG dtls TRACE: 06:55:18.883242 handshaker.go:151: [handshake:server] Flight 0: Preparing
  WhbypG dtls TRACE: 06:55:18.883291 handshaker.go:151: [handshake:server] Flight 0: Sending
  WhbypG dtls TRACE: 06:55:18.883297 handshaker.go:151: [handshake:server] Flight 0: Waiting
  WhbypG dtls TRACE: 06:55:18.883342 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  WhbypG dtls TRACE: 06:55:18.883346 handshaker.go:151: [handshake:server] Flight 2: Preparing
  gRUVWl dtls TRACE: 06:55:18.883341 handshaker.go:151: [handshake:server] Flight 0: Preparing
  gRUVWl dtls TRACE: 06:55:18.883361 handshaker.go:151: [handshake:server] Flight 0: Sending
  pRD23F dtls TRACE: 06:55:18.883361 handshaker.go:151: [handshake:server] Flight 0: Preparing
  XRzrMj dtls TRACE: 06:55:18.883371 handshaker.go:151: [handshake:server] Flight 0: Preparing
  XRzrMj dtls TRACE: 06:55:18.883378 handshaker.go:151: [handshake:server] Flight 0: Sending
  XRzrMj dtls TRACE: 06:55:18.883381 handshaker.go:151: [handshake:server] Flight 0: Waiting
  XRzrMj dtls TRACE: 06:55:18.883425 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  XRzrMj dtls TRACE: 06:55:18.883429 handshaker.go:151: [handshake:server] Flight 2: Preparing
  XRzrMj dtls TRACE: 06:55:18.883433 handshaker.go:151: [handshake:server] Flight 2: Sending
  XRzrMj dtls TRACE: 06:55:18.883438 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  pRD23F dtls TRACE: 06:55:18.883372 handshaker.go:151: [handshake:server] Flight 0: Sending
  WhbypG dtls TRACE: 06:55:18.883350 handshaker.go:151: [handshake:server] Flight 2: Sending
  gRUVWl dtls TRACE: 06:55:18.883365 handshaker.go:151: [handshake:server] Flight 0: Waiting
  WhbypG dtls TRACE: 06:55:18.883462 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  WhbypG dtls TRACE: 06:55:18.883483 handshaker.go:151: [handshake:server] Flight 2: Waiting
  L4lIOf dtls TRACE: 06:55:18.883371 handshaker.go:151: [handshake:server] Flight 0: Preparing
  L4lIOf dtls TRACE: 06:55:18.883493 handshaker.go:151: [handshake:server] Flight 0: Sending
  L4lIOf dtls TRACE: 06:55:18.883496 handshaker.go:151: [handshake:server] Flight 0: Waiting
  gRUVWl dtls TRACE: 06:55:18.883502 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  gRUVWl dtls TRACE: 06:55:18.883507 handshaker.go:151: [handshake:server] Flight 2: Preparing
  gRUVWl dtls TRACE: 06:55:18.883512 handshaker.go:151: [handshake:server] Flight 2: Sending
  gRUVWl dtls TRACE: 06:55:18.883516 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  XRzrMj dtls TRACE: 06:55:18.883457 handshaker.go:151: [handshake:server] Flight 2: Waiting
  pRD23F dtls TRACE: 06:55:18.883457 handshaker.go:151: [handshake:server] Flight 0: Waiting
  gRUVWl dtls TRACE: 06:55:18.883531 handshaker.go:151: [handshake:server] Flight 2: Waiting
  8YrRCj dtls TRACE: 06:55:18.883453 handshaker.go:151: [handshake:server] Flight 0: Preparing
  8YrRCj dtls TRACE: 06:55:18.883543 handshaker.go:151: [handshake:server] Flight 0: Sending
  8YrRCj dtls TRACE: 06:55:18.883546 handshaker.go:151: [handshake:server] Flight 0: Waiting
  pRD23F dtls TRACE: 06:55:18.883571 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  pRD23F dtls TRACE: 06:55:18.883576 handshaker.go:151: [handshake:server] Flight 2: Preparing
  pRD23F dtls TRACE: 06:55:18.883580 handshaker.go:151: [handshake:server] Flight 2: Sending
  pRD23F dtls TRACE: 06:55:18.883584 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  8YrRCj dtls TRACE: 06:55:18.883587 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  8YrRCj dtls TRACE: 06:55:18.883590 handshaker.go:151: [handshake:server] Flight 2: Preparing
  8YrRCj dtls TRACE: 06:55:18.883593 handshaker.go:151: [handshake:server] Flight 2: Sending
  8YrRCj dtls TRACE: 06:55:18.883597 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  pRD23F dtls TRACE: 06:55:18.883602 handshaker.go:151: [handshake:server] Flight 2: Waiting
  8YrRCj dtls TRACE: 06:55:18.883609 handshaker.go:151: [handshake:server] Flight 2: Waiting
  L4lIOf dtls TRACE: 06:55:18.883538 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  L4lIOf dtls TRACE: 06:55:18.883616 handshaker.go:151: [handshake:server] Flight 2: Preparing
  L4lIOf dtls TRACE: 06:55:18.883619 handshaker.go:151: [handshake:server] Flight 2: Sending
  L4lIOf dtls TRACE: 06:55:18.883623 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  L4lIOf dtls TRACE: 06:55:18.883636 handshaker.go:151: [handshake:server] Flight 2: Waiting
  pRD23F dtls TRACE: 06:55:18.971298 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  pRD23F dtls TRACE: 06:55:18.971309 handshaker.go:151: [handshake:server] Flight 4: Preparing
  pRD23F dtls TRACE: 06:55:18.971319 handshaker.go:151: [handshake:server] Flight 4: Sending
  pRD23F dtls TRACE: 06:55:18.971329 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  pRD23F dtls TRACE: 06:55:18.971337 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  pRD23F dtls TRACE: 06:55:18.971342 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  pRD23F dtls TRACE: 06:55:18.971368 handshaker.go:151: [handshake:server] Flight 4: Waiting
  8YrRCj dtls TRACE: 06:55:18.973331 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  8YrRCj dtls TRACE: 06:55:18.973342 handshaker.go:151: [handshake:server] Flight 4: Preparing
  8YrRCj dtls TRACE: 06:55:18.973348 handshaker.go:151: [handshake:server] Flight 4: Sending
  8YrRCj dtls TRACE: 06:55:18.973372 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  8YrRCj dtls TRACE: 06:55:18.973380 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  8YrRCj dtls TRACE: 06:55:18.973385 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  8YrRCj dtls TRACE: 06:55:18.973411 handshaker.go:151: [handshake:server] Flight 4: Waiting
  gRUVWl dtls TRACE: 06:55:18.987392 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  gRUVWl dtls TRACE: 06:55:18.987403 handshaker.go:151: [handshake:server] Flight 4: Preparing
  gRUVWl dtls TRACE: 06:55:18.987410 handshaker.go:151: [handshake:server] Flight 4: Sending
  gRUVWl dtls TRACE: 06:55:18.987417 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  gRUVWl dtls TRACE: 06:55:18.987423 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  gRUVWl dtls TRACE: 06:55:18.987432 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  gRUVWl dtls TRACE: 06:55:18.987462 handshaker.go:151: [handshake:server] Flight 4: Waiting
  XRzrMj dtls TRACE: 06:55:19.016107 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  XRzrMj dtls TRACE: 06:55:19.016119 handshaker.go:151: [handshake:server] Flight 4: Preparing
  XRzrMj dtls TRACE: 06:55:19.016126 handshaker.go:151: [handshake:server] Flight 4: Sending
  XRzrMj dtls TRACE: 06:55:19.016133 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  XRzrMj dtls TRACE: 06:55:19.016140 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  XRzrMj dtls TRACE: 06:55:19.016145 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  XRzrMj dtls TRACE: 06:55:19.016165 handshaker.go:151: [handshake:server] Flight 4: Waiting
  8YrRCj dtls DEBUG: 06:55:19.031127 conn.go:722: CipherSuite not initialized, queuing packet
  8YrRCj dtls DEBUG: 06:55:19.031141 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  8YrRCj dtls TRACE: 06:55:19.031215 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  8YrRCj dtls TRACE: 06:55:19.031255 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  8YrRCj dtls TRACE: 06:55:19.031277 handshaker.go:151: [handshake:server] Flight 6: Preparing
  8YrRCj dtls TRACE: 06:55:19.031291 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  8YrRCj dtls TRACE: 06:55:19.031297 handshaker.go:151: [handshake:server] Flight 6: Sending
  8YrRCj dtls TRACE: 06:55:19.031305 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  8YrRCj dtls TRACE: 06:55:19.031337 handshaker.go:151: [handshake:server] Flight 6: Finished
  8YrRCj dtls TRACE: 06:55:19.031440 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:36103
  Mnyldx dtls TRACE: 06:55:19.031505 handshaker.go:151: [handshake:server] Flight 0: Preparing
  Mnyldx dtls TRACE: 06:55:19.031543 handshaker.go:151: [handshake:server] Flight 0: Sending
  Mnyldx dtls TRACE: 06:55:19.031565 handshaker.go:151: [handshake:server] Flight 0: Waiting
  Mnyldx dtls TRACE: 06:55:19.031615 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  Mnyldx dtls TRACE: 06:55:19.031681 handshaker.go:151: [handshake:server] Flight 2: Preparing
  Mnyldx dtls TRACE: 06:55:19.031688 handshaker.go:151: [handshake:server] Flight 2: Sending
  Mnyldx dtls TRACE: 06:55:19.031695 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  Mnyldx dtls TRACE: 06:55:19.031725 handshaker.go:151: [handshake:server] Flight 2: Waiting
  L4lIOf dtls TRACE: 06:55:19.062366 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  L4lIOf dtls TRACE: 06:55:19.062378 handshaker.go:151: [handshake:server] Flight 4: Preparing
  L4lIOf dtls TRACE: 06:55:19.062385 handshaker.go:151: [handshake:server] Flight 4: Sending
  L4lIOf dtls TRACE: 06:55:19.062392 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  L4lIOf dtls TRACE: 06:55:19.062397 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  L4lIOf dtls TRACE: 06:55:19.062401 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  L4lIOf dtls TRACE: 06:55:19.062428 handshaker.go:151: [handshake:server] Flight 4: Waiting
  WhbypG dtls TRACE: 06:55:19.131108 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  WhbypG dtls TRACE: 06:55:19.131150 handshaker.go:151: [handshake:server] Flight 4: Preparing
  WhbypG dtls TRACE: 06:55:19.131185 handshaker.go:151: [handshake:server] Flight 4: Sending
  WhbypG dtls TRACE: 06:55:19.131213 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  WhbypG dtls TRACE: 06:55:19.131243 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  WhbypG dtls TRACE: 06:55:19.131274 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  WhbypG dtls TRACE: 06:55:19.131366 handshaker.go:151: [handshake:server] Flight 4: Waiting
  gRUVWl dtls DEBUG: 06:55:19.177052 conn.go:722: CipherSuite not initialized, queuing packet
  gRUVWl dtls DEBUG: 06:55:19.177098 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  gRUVWl dtls TRACE: 06:55:19.177240 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  gRUVWl dtls TRACE: 06:55:19.177302 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  gRUVWl dtls TRACE: 06:55:19.177325 handshaker.go:151: [handshake:server] Flight 6: Preparing
  gRUVWl dtls TRACE: 06:55:19.177379 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  gRUVWl dtls TRACE: 06:55:19.177407 handshaker.go:151: [handshake:server] Flight 6: Sending
  gRUVWl dtls TRACE: 06:55:19.177434 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  gRUVWl dtls TRACE: 06:55:19.177554 handshaker.go:151: [handshake:server] Flight 6: Finished
  gRUVWl dtls TRACE: 06:55:19.177584 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:60902
  pRD23F dtls DEBUG: 06:55:19.177836 conn.go:722: CipherSuite not initialized, queuing packet
  pRD23F dtls DEBUG: 06:55:19.177872 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  pRD23F dtls TRACE: 06:55:19.178027 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  pRD23F dtls TRACE: 06:55:19.178075 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  pRD23F dtls TRACE: 06:55:19.178094 handshaker.go:151: [handshake:server] Flight 6: Preparing
  pRD23F dtls TRACE: 06:55:19.178137 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  pRD23F dtls TRACE: 06:55:19.178158 handshaker.go:151: [handshake:server] Flight 6: Sending
  pRD23F dtls TRACE: 06:55:19.178179 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  pRD23F dtls TRACE: 06:55:19.178279 handshaker.go:151: [handshake:server] Flight 6: Finished
  pRD23F dtls TRACE: 06:55:19.178308 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:36835
  TPlQUX dtls TRACE: 06:55:19.178464 handshaker.go:151: [handshake:server] Flight 0: Preparing
  TPlQUX dtls TRACE: 06:55:19.178498 handshaker.go:151: [handshake:server] Flight 0: Sending
  TPlQUX dtls TRACE: 06:55:19.178519 handshaker.go:151: [handshake:server] Flight 0: Waiting
  TPlQUX dtls TRACE: 06:55:19.178734 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  TPlQUX dtls TRACE: 06:55:19.178752 handshaker.go:151: [handshake:server] Flight 2: Preparing
  TPlQUX dtls TRACE: 06:55:19.178771 handshaker.go:151: [handshake:server] Flight 2: Sending
  TPlQUX dtls TRACE: 06:55:19.178799 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  TPlQUX dtls TRACE: 06:55:19.178883 handshaker.go:151: [handshake:server] Flight 2: Waiting
  0PPM0n dtls TRACE: 06:55:19.178935 handshaker.go:151: [handshake:server] Flight 0: Preparing
  0PPM0n dtls TRACE: 06:55:19.178963 handshaker.go:151: [handshake:server] Flight 0: Sending
  0PPM0n dtls TRACE: 06:55:19.178979 handshaker.go:151: [handshake:server] Flight 0: Waiting
  0PPM0n dtls TRACE: 06:55:19.179193 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  0PPM0n dtls TRACE: 06:55:19.179210 handshaker.go:151: [handshake:server] Flight 2: Preparing
  0PPM0n dtls TRACE: 06:55:19.179227 handshaker.go:151: [handshake:server] Flight 2: Sending
  0PPM0n dtls TRACE: 06:55:19.179246 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  0PPM0n dtls TRACE: 06:55:19.179309 handshaker.go:151: [handshake:server] Flight 2: Waiting
  XRzrMj dtls DEBUG: 06:55:19.203086 conn.go:722: CipherSuite not initialized, queuing packet
  XRzrMj dtls DEBUG: 06:55:19.204023 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  XRzrMj dtls TRACE: 06:55:19.204294 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  XRzrMj dtls TRACE: 06:55:19.204421 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  XRzrMj dtls TRACE: 06:55:19.204488 handshaker.go:151: [handshake:server] Flight 6: Preparing
  XRzrMj dtls TRACE: 06:55:19.204520 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  XRzrMj dtls TRACE: 06:55:19.204536 handshaker.go:151: [handshake:server] Flight 6: Sending
  XRzrMj dtls TRACE: 06:55:19.204583 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  XRzrMj dtls TRACE: 06:55:19.204765 handshaker.go:151: [handshake:server] Flight 6: Finished
  XRzrMj dtls TRACE: 06:55:19.204791 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:52727
  E6DCrU dtls TRACE: 06:55:19.204989 handshaker.go:151: [handshake:server] Flight 0: Preparing
  E6DCrU dtls TRACE: 06:55:19.205008 handshaker.go:151: [handshake:server] Flight 0: Sending
  E6DCrU dtls TRACE: 06:55:19.205017 handshaker.go:151: [handshake:server] Flight 0: Waiting
  E6DCrU dtls TRACE: 06:55:19.205278 handshaker.go:274: [handshake:server] Flight 0 -> Flight 2
  E6DCrU dtls TRACE: 06:55:19.205332 handshaker.go:151: [handshake:server] Flight 2: Preparing
  E6DCrU dtls TRACE: 06:55:19.205345 handshaker.go:151: [handshake:server] Flight 2: Sending
  E6DCrU dtls TRACE: 06:55:19.205357 conn.go:372: [handshake:server] -> HelloVerifyRequest (epoch: 0, seq: 0)
  E6DCrU dtls TRACE: 06:55:19.205474 handshaker.go:151: [handshake:server] Flight 2: Waiting
  Mnyldx dtls TRACE: 06:55:19.209376 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  Mnyldx dtls TRACE: 06:55:19.209419 handshaker.go:151: [handshake:server] Flight 4: Preparing
  Mnyldx dtls TRACE: 06:55:19.209435 handshaker.go:151: [handshake:server] Flight 4: Sending
  Mnyldx dtls TRACE: 06:55:19.209451 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  Mnyldx dtls TRACE: 06:55:19.209468 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  Mnyldx dtls TRACE: 06:55:19.209496 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  Mnyldx dtls TRACE: 06:55:19.209538 handshaker.go:151: [handshake:server] Flight 4: Waiting
  gRUVWl dtls TRACE: 06:55:19.236125 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:60902
  TPlQUX dtls TRACE: 06:55:19.243656 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  TPlQUX dtls TRACE: 06:55:19.243675 handshaker.go:151: [handshake:server] Flight 4: Preparing
  TPlQUX dtls TRACE: 06:55:19.244112 handshaker.go:151: [handshake:server] Flight 4: Sending
  TPlQUX dtls TRACE: 06:55:19.244135 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  TPlQUX dtls TRACE: 06:55:19.244153 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  TPlQUX dtls TRACE: 06:55:19.244168 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  TPlQUX dtls TRACE: 06:55:19.244221 handshaker.go:151: [handshake:server] Flight 4: Waiting
  8YrRCj dtls TRACE: 06:55:19.270952 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:36103
  pRD23F dtls TRACE: 06:55:19.283766 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:36835
  L4lIOf dtls DEBUG: 06:55:19.295079 conn.go:722: CipherSuite not initialized, queuing packet
  L4lIOf dtls DEBUG: 06:55:19.295092 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  L4lIOf dtls TRACE: 06:55:19.295133 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  L4lIOf dtls TRACE: 06:55:19.295152 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  L4lIOf dtls TRACE: 06:55:19.295157 handshaker.go:151: [handshake:server] Flight 6: Preparing
  L4lIOf dtls TRACE: 06:55:19.295170 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  L4lIOf dtls TRACE: 06:55:19.295179 handshaker.go:151: [handshake:server] Flight 6: Sending
  L4lIOf dtls TRACE: 06:55:19.295185 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  L4lIOf dtls TRACE: 06:55:19.295215 handshaker.go:151: [handshake:server] Flight 6: Finished
  L4lIOf dtls TRACE: 06:55:19.295223 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:45003
  WhbypG dtls DEBUG: 06:55:19.312070 conn.go:722: CipherSuite not initialized, queuing packet
  WhbypG dtls DEBUG: 06:55:19.312083 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  WhbypG dtls TRACE: 06:55:19.312118 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  WhbypG dtls TRACE: 06:55:19.312130 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  WhbypG dtls TRACE: 06:55:19.312135 handshaker.go:151: [handshake:server] Flight 6: Preparing
  WhbypG dtls TRACE: 06:55:19.312149 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  WhbypG dtls TRACE: 06:55:19.312153 handshaker.go:151: [handshake:server] Flight 6: Sending
  WhbypG dtls TRACE: 06:55:19.312159 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  WhbypG dtls TRACE: 06:55:19.312187 handshaker.go:151: [handshake:server] Flight 6: Finished
  WhbypG dtls TRACE: 06:55:19.312194 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:56450
  Mnyldx dtls DEBUG: 06:55:19.332383 conn.go:722: CipherSuite not initialized, queuing packet
  Mnyldx dtls DEBUG: 06:55:19.332396 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  Mnyldx dtls TRACE: 06:55:19.332435 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  Mnyldx dtls TRACE: 06:55:19.332448 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  Mnyldx dtls TRACE: 06:55:19.332453 handshaker.go:151: [handshake:server] Flight 6: Preparing
  Mnyldx dtls TRACE: 06:55:19.332463 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  Mnyldx dtls TRACE: 06:55:19.332467 handshaker.go:151: [handshake:server] Flight 6: Sending
  Mnyldx dtls TRACE: 06:55:19.332473 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  Mnyldx dtls TRACE: 06:55:19.332506 handshaker.go:151: [handshake:server] Flight 6: Finished
  Mnyldx dtls TRACE: 06:55:19.332514 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:50387
  0PPM0n dtls TRACE: 06:55:19.338540 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  0PPM0n dtls TRACE: 06:55:19.338578 handshaker.go:151: [handshake:server] Flight 4: Preparing
  0PPM0n dtls TRACE: 06:55:19.338587 handshaker.go:151: [handshake:server] Flight 4: Sending
  0PPM0n dtls TRACE: 06:55:19.338597 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  0PPM0n dtls TRACE: 06:55:19.338606 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  0PPM0n dtls TRACE: 06:55:19.338611 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  0PPM0n dtls TRACE: 06:55:19.338632 handshaker.go:151: [handshake:server] Flight 4: Waiting
  E6DCrU dtls TRACE: 06:55:19.373048 handshaker.go:274: [handshake:server] Flight 2 -> Flight 4
  E6DCrU dtls TRACE: 06:55:19.373319 handshaker.go:151: [handshake:server] Flight 4: Preparing
  E6DCrU dtls TRACE: 06:55:19.373329 handshaker.go:151: [handshake:server] Flight 4: Sending
  E6DCrU dtls TRACE: 06:55:19.373336 conn.go:372: [handshake:server] -> ServerHello (epoch: 0, seq: 1)
  E6DCrU dtls TRACE: 06:55:19.373343 conn.go:372: [handshake:server] -> ServerKeyExchange (epoch: 0, seq: 2)
  E6DCrU dtls TRACE: 06:55:19.373348 conn.go:372: [handshake:server] -> ServerHelloDone (epoch: 0, seq: 3)
  E6DCrU dtls TRACE: 06:55:19.373413 handshaker.go:151: [handshake:server] Flight 4: Waiting
  XRzrMj dtls TRACE: 06:55:19.400369 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:52727
  TPlQUX dtls DEBUG: 06:55:19.404634 conn.go:722: CipherSuite not initialized, queuing packet
  TPlQUX dtls DEBUG: 06:55:19.404646 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  TPlQUX dtls TRACE: 06:55:19.404683 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  TPlQUX dtls TRACE: 06:55:19.404699 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  TPlQUX dtls TRACE: 06:55:19.404738 handshaker.go:151: [handshake:server] Flight 6: Preparing
  TPlQUX dtls TRACE: 06:55:19.404756 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  TPlQUX dtls TRACE: 06:55:19.404761 handshaker.go:151: [handshake:server] Flight 6: Sending
  TPlQUX dtls TRACE: 06:55:19.404767 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  TPlQUX dtls TRACE: 06:55:19.404799 handshaker.go:151: [handshake:server] Flight 6: Finished
  TPlQUX dtls TRACE: 06:55:19.404806 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:44437
  WhbypG dtls TRACE: 06:55:19.446507 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:56450
  E6DCrU dtls DEBUG: 06:55:19.466928 conn.go:722: CipherSuite not initialized, queuing packet
  E6DCrU dtls DEBUG: 06:55:19.466943 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  E6DCrU dtls TRACE: 06:55:19.466986 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  E6DCrU dtls TRACE: 06:55:19.467001 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  E6DCrU dtls TRACE: 06:55:19.467006 handshaker.go:151: [handshake:server] Flight 6: Preparing
  E6DCrU dtls TRACE: 06:55:19.467016 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  E6DCrU dtls TRACE: 06:55:19.467021 handshaker.go:151: [handshake:server] Flight 6: Sending
  E6DCrU dtls TRACE: 06:55:19.467027 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  E6DCrU dtls TRACE: 06:55:19.467056 handshaker.go:151: [handshake:server] Flight 6: Finished
  E6DCrU dtls TRACE: 06:55:19.467070 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:45184
  Mnyldx dtls TRACE: 06:55:19.494462 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:50387
  L4lIOf dtls TRACE: 06:55:19.528928 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:45003
  0PPM0n dtls DEBUG: 06:55:19.529985 conn.go:722: CipherSuite not initialized, queuing packet
  0PPM0n dtls DEBUG: 06:55:19.530007 conn.go:644: received packet of next epoch, queuing packet
  Client's hint: Pion DTLS Server
  0PPM0n dtls TRACE: 06:55:19.530055 conn.go:728: server: <- ChangeCipherSpec (epoch: 1)
  0PPM0n dtls TRACE: 06:55:19.530070 handshaker.go:274: [handshake:server] Flight 4 -> Flight 6
  0PPM0n dtls TRACE: 06:55:19.530078 handshaker.go:151: [handshake:server] Flight 6: Preparing
  0PPM0n dtls TRACE: 06:55:19.530088 handshaker.go:226: [handshake:server] -> changeCipherSpec (epoch: 1)
  0PPM0n dtls TRACE: 06:55:19.530092 handshaker.go:151: [handshake:server] Flight 6: Sending
  0PPM0n dtls TRACE: 06:55:19.530098 conn.go:372: [handshake:server] -> Finished (epoch: 1, seq: 4)
  0PPM0n dtls TRACE: 06:55:19.530135 handshaker.go:151: [handshake:server] Flight 6: Finished
  0PPM0n dtls TRACE: 06:55:19.530143 conn.go:198: Handshake Completed
  Connected to 127.0.0.1:45125
  TPlQUX dtls TRACE: 06:55:19.550784 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:44437
  0PPM0n dtls TRACE: 06:55:19.616527 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:45125
  E6DCrU dtls TRACE: 06:55:19.648833 conn.go:710: server: <- Alert LevelWarning: CloseNotify
  Disconnected  127.0.0.1:45184
  ```

  </details>

  <details><summary>Client-side log with 5 concurrent listeners.</summary>

  ```shell
  $ mkdir -p build && go build -o build/ ./examples/dial/psk && for i in {1..10}; do ./build/psk <<< 'exit' &; done
  [2] 11788
  [3] 11789
  [4] 11790
  [5] 11791
  [6] 11792
  [7] 11793
  [8] 11794
  [9] 11795
  [10] 11796
  [11] 11797
  cMhhOl dtls TRACE: 06:55:18.882901 handshaker.go:151: [handshake:client] Flight 1: Preparing
  dIMCe1 dtls TRACE: 06:55:18.882950 handshaker.go:151: [handshake:client] Flight 1: Preparing
  KkCtzC dtls TRACE: 06:55:18.882957 handshaker.go:151: [handshake:client] Flight 1: Preparing
  IcPwE7 dtls TRACE: 06:55:18.882962 handshaker.go:151: [handshake:client] Flight 1: Preparing
  WpckXj dtls TRACE: 06:55:18.882909 handshaker.go:151: [handshake:client] Flight 1: Preparing
  BpQksA dtls TRACE: 06:55:18.883003 handshaker.go:151: [handshake:client] Flight 1: Preparing
  VkBrfh dtls TRACE: 06:55:18.883013 handshaker.go:151: [handshake:client] Flight 1: Preparing
  BpQksA dtls TRACE: 06:55:18.883039 handshaker.go:151: [handshake:client] Flight 1: Sending
  VkBrfh dtls TRACE: 06:55:18.883049 handshaker.go:151: [handshake:client] Flight 1: Sending
  BpQksA dtls TRACE: 06:55:18.883059 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  VkBrfh dtls TRACE: 06:55:18.883061 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  VkBrfh dtls TRACE: 06:55:18.883102 handshaker.go:151: [handshake:client] Flight 1: Waiting
  BpQksA dtls TRACE: 06:55:18.883102 handshaker.go:151: [handshake:client] Flight 1: Waiting
  KkCtzC dtls TRACE: 06:55:18.883129 handshaker.go:151: [handshake:client] Flight 1: Sending
  IcPwE7 dtls TRACE: 06:55:18.883136 handshaker.go:151: [handshake:client] Flight 1: Sending
  KkCtzC dtls TRACE: 06:55:18.883143 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  IcPwE7 dtls TRACE: 06:55:18.883149 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  KkCtzC dtls TRACE: 06:55:18.883164 handshaker.go:151: [handshake:client] Flight 1: Waiting
  IcPwE7 dtls TRACE: 06:55:18.883178 handshaker.go:151: [handshake:client] Flight 1: Waiting
  dIMCe1 dtls TRACE: 06:55:18.883208 handshaker.go:151: [handshake:client] Flight 1: Sending
  dIMCe1 dtls TRACE: 06:55:18.883221 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  dIMCe1 dtls TRACE: 06:55:18.883241 handshaker.go:151: [handshake:client] Flight 1: Waiting
  cMhhOl dtls TRACE: 06:55:18.883388 handshaker.go:151: [handshake:client] Flight 1: Sending
  cMhhOl dtls TRACE: 06:55:18.883404 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  cMhhOl dtls TRACE: 06:55:18.883432 handshaker.go:151: [handshake:client] Flight 1: Waiting
  KkCtzC dtls TRACE: 06:55:18.883584 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  KkCtzC dtls TRACE: 06:55:18.883656 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 132ms
  WpckXj dtls TRACE: 06:55:18.883836 handshaker.go:151: [handshake:client] Flight 1: Sending
  BpQksA dtls TRACE: 06:55:18.883636 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  BpQksA dtls TRACE: 06:55:18.883953 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 103ms
  dIMCe1 dtls TRACE: 06:55:18.883697 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  dIMCe1 dtls TRACE: 06:55:18.883997 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 87ms
  cMhhOl dtls TRACE: 06:55:18.883883 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  cMhhOl dtls TRACE: 06:55:18.884029 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 89ms
  IcPwE7 dtls TRACE: 06:55:18.884067 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  WpckXj dtls TRACE: 06:55:18.883934 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  IcPwE7 dtls TRACE: 06:55:18.884083 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 178ms
  Gh42hw dtls TRACE: 06:55:18.885282 handshaker.go:151: [handshake:client] Flight 1: Preparing
  Gh42hw dtls TRACE: 06:55:18.885339 handshaker.go:151: [handshake:client] Flight 1: Sending
  Gh42hw dtls TRACE: 06:55:18.885352 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  Gh42hw dtls TRACE: 06:55:18.885389 handshaker.go:151: [handshake:client] Flight 1: Waiting
  nTjWbB dtls TRACE: 06:55:18.887301 handshaker.go:151: [handshake:client] Flight 1: Preparing
  nTjWbB dtls TRACE: 06:55:18.887345 handshaker.go:151: [handshake:client] Flight 1: Sending
  nTjWbB dtls TRACE: 06:55:18.887358 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  nTjWbB dtls TRACE: 06:55:18.887396 handshaker.go:151: [handshake:client] Flight 1: Waiting
  WpckXj dtls TRACE: 06:55:18.888073 handshaker.go:151: [handshake:client] Flight 1: Waiting
  VkBrfh dtls TRACE: 06:55:18.888130 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  VkBrfh dtls TRACE: 06:55:18.888140 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 242ms
  Akl6hf dtls TRACE: 06:55:18.888180 handshaker.go:151: [handshake:client] Flight 1: Preparing
  Akl6hf dtls TRACE: 06:55:18.888227 handshaker.go:151: [handshake:client] Flight 1: Sending
  Akl6hf dtls TRACE: 06:55:18.888241 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 0)
  Akl6hf dtls TRACE: 06:55:18.888321 handshaker.go:151: [handshake:client] Flight 1: Waiting
  dIMCe1 dtls TRACE: 06:55:18.971134 handshaker.go:151: [handshake:client] Flight 3: Preparing
  dIMCe1 dtls TRACE: 06:55:18.971150 handshaker.go:151: [handshake:client] Flight 3: Sending
  dIMCe1 dtls TRACE: 06:55:18.971174 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  dIMCe1 dtls TRACE: 06:55:18.971409 handshaker.go:151: [handshake:client] Flight 3: Waiting
  dIMCe1 dtls TRACE: 06:55:18.971425 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  dIMCe1 dtls TRACE: 06:55:18.971433 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  dIMCe1 dtls TRACE: 06:55:18.971439 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 203ms
  cMhhOl dtls TRACE: 06:55:18.973203 handshaker.go:151: [handshake:client] Flight 3: Preparing
  cMhhOl dtls TRACE: 06:55:18.973223 handshaker.go:151: [handshake:client] Flight 3: Sending
  cMhhOl dtls TRACE: 06:55:18.973241 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  cMhhOl dtls TRACE: 06:55:18.973284 handshaker.go:151: [handshake:client] Flight 3: Waiting
  cMhhOl dtls TRACE: 06:55:18.973594 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  cMhhOl dtls TRACE: 06:55:18.973791 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  cMhhOl dtls TRACE: 06:55:18.973802 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 57ms
  BpQksA dtls TRACE: 06:55:18.987118 handshaker.go:151: [handshake:client] Flight 3: Preparing
  BpQksA dtls TRACE: 06:55:18.987140 handshaker.go:151: [handshake:client] Flight 3: Sending
  BpQksA dtls TRACE: 06:55:18.987157 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  BpQksA dtls TRACE: 06:55:18.987201 handshaker.go:151: [handshake:client] Flight 3: Waiting
  BpQksA dtls TRACE: 06:55:18.987502 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  BpQksA dtls TRACE: 06:55:18.987513 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  BpQksA dtls TRACE: 06:55:18.987598 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 188ms
  KkCtzC dtls TRACE: 06:55:19.015931 handshaker.go:151: [handshake:client] Flight 3: Preparing
  KkCtzC dtls TRACE: 06:55:19.015957 handshaker.go:151: [handshake:client] Flight 3: Sending
  KkCtzC dtls TRACE: 06:55:19.015976 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  KkCtzC dtls TRACE: 06:55:19.016032 handshaker.go:151: [handshake:client] Flight 3: Waiting
  KkCtzC dtls TRACE: 06:55:19.016199 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  KkCtzC dtls TRACE: 06:55:19.016229 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  KkCtzC dtls TRACE: 06:55:19.016252 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 186ms
  cMhhOl dtls TRACE: 06:55:19.030935 handshaker.go:151: [handshake:client] Flight 5: Preparing
  cMhhOl dtls TRACE: 06:55:19.031001 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  cMhhOl dtls TRACE: 06:55:19.031007 handshaker.go:151: [handshake:client] Flight 5: Sending
  cMhhOl dtls TRACE: 06:55:19.031012 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  cMhhOl dtls TRACE: 06:55:19.031020 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  cMhhOl dtls TRACE: 06:55:19.031084 handshaker.go:151: [handshake:client] Flight 5: Waiting
  cMhhOl dtls TRACE: 06:55:19.031367 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  cMhhOl dtls TRACE: 06:55:19.031420 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  cMhhOl dtls TRACE: 06:55:19.031532 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 239ms
  Gh42hw dtls TRACE: 06:55:19.031794 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  Gh42hw dtls TRACE: 06:55:19.031867 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 177ms
  IcPwE7 dtls TRACE: 06:55:19.062213 handshaker.go:151: [handshake:client] Flight 3: Preparing
  IcPwE7 dtls TRACE: 06:55:19.062230 handshaker.go:151: [handshake:client] Flight 3: Sending
  IcPwE7 dtls TRACE: 06:55:19.062260 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  IcPwE7 dtls TRACE: 06:55:19.062307 handshaker.go:151: [handshake:client] Flight 3: Waiting
  IcPwE7 dtls TRACE: 06:55:19.062463 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  IcPwE7 dtls TRACE: 06:55:19.062763 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  IcPwE7 dtls TRACE: 06:55:19.062775 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 232ms
  VkBrfh dtls TRACE: 06:55:19.130439 handshaker.go:151: [handshake:client] Flight 3: Preparing
  VkBrfh dtls TRACE: 06:55:19.130563 handshaker.go:151: [handshake:client] Flight 3: Sending
  VkBrfh dtls TRACE: 06:55:19.130667 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  VkBrfh dtls TRACE: 06:55:19.130988 handshaker.go:151: [handshake:client] Flight 3: Waiting
  VkBrfh dtls TRACE: 06:55:19.131572 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  VkBrfh dtls TRACE: 06:55:19.131696 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  VkBrfh dtls TRACE: 06:55:19.131738 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 180ms
  dIMCe1 dtls TRACE: 06:55:19.176194 handshaker.go:151: [handshake:client] Flight 5: Preparing
  BpQksA dtls TRACE: 06:55:19.176216 handshaker.go:151: [handshake:client] Flight 5: Preparing
  dIMCe1 dtls TRACE: 06:55:19.176391 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  dIMCe1 dtls TRACE: 06:55:19.176416 handshaker.go:151: [handshake:client] Flight 5: Sending
  dIMCe1 dtls TRACE: 06:55:19.176436 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  dIMCe1 dtls TRACE: 06:55:19.176478 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  BpQksA dtls TRACE: 06:55:19.176548 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  BpQksA dtls TRACE: 06:55:19.176577 handshaker.go:151: [handshake:client] Flight 5: Sending
  BpQksA dtls TRACE: 06:55:19.176599 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  BpQksA dtls TRACE: 06:55:19.176629 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  dIMCe1 dtls TRACE: 06:55:19.176644 handshaker.go:151: [handshake:client] Flight 5: Waiting
  BpQksA dtls TRACE: 06:55:19.176814 handshaker.go:151: [handshake:client] Flight 5: Waiting
  BpQksA dtls TRACE: 06:55:19.177982 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  nTjWbB dtls TRACE: 06:55:19.179367 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  nTjWbB dtls TRACE: 06:55:19.179442 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 63ms
  WpckXj dtls TRACE: 06:55:19.179603 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  WpckXj dtls TRACE: 06:55:19.180177 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 158ms
  dIMCe1 dtls TRACE: 06:55:19.179789 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  dIMCe1 dtls TRACE: 06:55:19.180488 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  BpQksA dtls TRACE: 06:55:19.178470 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  BpQksA dtls TRACE: 06:55:19.180677 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 55ms
  dIMCe1 dtls TRACE: 06:55:19.180608 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 102ms
  KkCtzC dtls TRACE: 06:55:19.202565 handshaker.go:151: [handshake:client] Flight 5: Preparing
  KkCtzC dtls TRACE: 06:55:19.202766 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  KkCtzC dtls TRACE: 06:55:19.202795 handshaker.go:151: [handshake:client] Flight 5: Sending
  KkCtzC dtls TRACE: 06:55:19.202814 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  KkCtzC dtls TRACE: 06:55:19.202837 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  KkCtzC dtls TRACE: 06:55:19.202956 handshaker.go:151: [handshake:client] Flight 5: Waiting
  KkCtzC dtls TRACE: 06:55:19.204779 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  KkCtzC dtls TRACE: 06:55:19.204966 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  KkCtzC dtls TRACE: 06:55:19.205068 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 195ms
  Akl6hf dtls TRACE: 06:55:19.205660 handshaker.go:274: [handshake:client] Flight 1 -> Flight 3
  Akl6hf dtls TRACE: 06:55:19.205771 handshaker.go:165: [handshake:client] Flight 3: Preparing but sleeping for 167ms
  Gh42hw dtls TRACE: 06:55:19.209001 handshaker.go:151: [handshake:client] Flight 3: Preparing
  Gh42hw dtls TRACE: 06:55:19.209047 handshaker.go:151: [handshake:client] Flight 3: Sending
  Gh42hw dtls TRACE: 06:55:19.209093 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  Gh42hw dtls TRACE: 06:55:19.209166 handshaker.go:151: [handshake:client] Flight 3: Waiting
  Gh42hw dtls TRACE: 06:55:19.209673 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  Gh42hw dtls TRACE: 06:55:19.209730 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  Gh42hw dtls TRACE: 06:55:19.210078 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 122ms
  BpQksA dtls TRACE: 06:55:19.235893 handshaker.go:151: [handshake:client] Flight 5: Finished
  BpQksA dtls TRACE: 06:55:19.235923 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF

  [10]  - 11796 done       ./build/psk <<< 'exit'
  nTjWbB dtls TRACE: 06:55:19.243410 handshaker.go:151: [handshake:client] Flight 3: Preparing
  nTjWbB dtls TRACE: 06:55:19.243439 handshaker.go:151: [handshake:client] Flight 3: Sending
  nTjWbB dtls TRACE: 06:55:19.243501 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  nTjWbB dtls TRACE: 06:55:19.243580 handshaker.go:151: [handshake:client] Flight 3: Waiting
  nTjWbB dtls TRACE: 06:55:19.244287 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  nTjWbB dtls TRACE: 06:55:19.244328 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  nTjWbB dtls TRACE: 06:55:19.244345 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 160ms
  cMhhOl dtls TRACE: 06:55:19.270634 handshaker.go:151: [handshake:client] Flight 5: Finished
  cMhhOl dtls TRACE: 06:55:19.270659 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF
  panic: EOF

  goroutine 20 [running]:
  github.com/pion/dtls/v2/examples/util.Check(0x652d00, 0xc000012120)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:65 +0x1b4
  github.com/pion/dtls/v2/examples/util.Chat.func1(0x654180, 0xc00011e580)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:31 +0x152
  created by github.com/pion/dtls/v2/examples/util.Chat
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:26 +0x64
  [3]    11789 exit 2     ./build/psk <<< 'exit'
  dIMCe1 dtls TRACE: 06:55:19.283588 handshaker.go:151: [handshake:client] Flight 5: Finished
  dIMCe1 dtls TRACE: 06:55:19.283615 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF
  panic: EOF

  goroutine 20 [running]:
  github.com/pion/dtls/v2/examples/util.Check(0x652d00, 0xc000012120)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:65 +0x1b4
  github.com/pion/dtls/v2/examples/util.Chat.func1(0x654180, 0xc00011e580)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:31 +0x152
  created by github.com/pion/dtls/v2/examples/util.Chat
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:26 +0x64

  [5]    11791 exit 2     ./build/psk <<< 'exit'
  IcPwE7 dtls TRACE: 06:55:19.294863 handshaker.go:151: [handshake:client] Flight 5: Preparing
  IcPwE7 dtls TRACE: 06:55:19.294944 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  IcPwE7 dtls TRACE: 06:55:19.294956 handshaker.go:151: [handshake:client] Flight 5: Sending
  IcPwE7 dtls TRACE: 06:55:19.294964 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  IcPwE7 dtls TRACE: 06:55:19.294974 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  IcPwE7 dtls TRACE: 06:55:19.295030 handshaker.go:151: [handshake:client] Flight 5: Waiting
  IcPwE7 dtls TRACE: 06:55:19.295445 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  IcPwE7 dtls TRACE: 06:55:19.295492 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  IcPwE7 dtls TRACE: 06:55:19.295646 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 233ms
  VkBrfh dtls TRACE: 06:55:19.311848 handshaker.go:151: [handshake:client] Flight 5: Preparing
  VkBrfh dtls TRACE: 06:55:19.311929 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  VkBrfh dtls TRACE: 06:55:19.311942 handshaker.go:151: [handshake:client] Flight 5: Sending
  VkBrfh dtls TRACE: 06:55:19.311948 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  VkBrfh dtls TRACE: 06:55:19.311955 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  VkBrfh dtls TRACE: 06:55:19.312007 handshaker.go:151: [handshake:client] Flight 5: Waiting
  VkBrfh dtls TRACE: 06:55:19.312211 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  VkBrfh dtls TRACE: 06:55:19.312245 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  VkBrfh dtls TRACE: 06:55:19.312256 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 134ms
  Gh42hw dtls TRACE: 06:55:19.332195 handshaker.go:151: [handshake:client] Flight 5: Preparing
  Gh42hw dtls TRACE: 06:55:19.332254 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  Gh42hw dtls TRACE: 06:55:19.332259 handshaker.go:151: [handshake:client] Flight 5: Sending
  Gh42hw dtls TRACE: 06:55:19.332264 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  Gh42hw dtls TRACE: 06:55:19.332270 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  Gh42hw dtls TRACE: 06:55:19.332333 handshaker.go:151: [handshake:client] Flight 5: Waiting
  Gh42hw dtls TRACE: 06:55:19.332541 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  Gh42hw dtls TRACE: 06:55:19.332787 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  Gh42hw dtls TRACE: 06:55:19.332995 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 161ms
  WpckXj dtls TRACE: 06:55:19.338381 handshaker.go:151: [handshake:client] Flight 3: Preparing
  WpckXj dtls TRACE: 06:55:19.338398 handshaker.go:151: [handshake:client] Flight 3: Sending
  WpckXj dtls TRACE: 06:55:19.338434 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  WpckXj dtls TRACE: 06:55:19.338476 handshaker.go:151: [handshake:client] Flight 3: Waiting
  WpckXj dtls TRACE: 06:55:19.338675 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  WpckXj dtls TRACE: 06:55:19.338695 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  WpckXj dtls TRACE: 06:55:19.338702 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 191ms
  Akl6hf dtls TRACE: 06:55:19.372898 handshaker.go:151: [handshake:client] Flight 3: Preparing
  Akl6hf dtls TRACE: 06:55:19.372915 handshaker.go:151: [handshake:client] Flight 3: Sending
  Akl6hf dtls TRACE: 06:55:19.372939 conn.go:372: [handshake:client] -> ClientHello (epoch: 0, seq: 1)
  Akl6hf dtls TRACE: 06:55:19.372984 handshaker.go:151: [handshake:client] Flight 3: Waiting
  Akl6hf dtls TRACE: 06:55:19.373467 flight3handler.go:78: [handshake] use cipher suite: TLS_PSK_WITH_AES_128_CCM_8
  Server's hint: Pion DTLS Client
  Akl6hf dtls TRACE: 06:55:19.373620 handshaker.go:274: [handshake:client] Flight 3 -> Flight 5
  Akl6hf dtls TRACE: 06:55:19.373632 handshaker.go:165: [handshake:client] Flight 5: Preparing but sleeping for 93ms
  KkCtzC dtls TRACE: 06:55:19.400214 handshaker.go:151: [handshake:client] Flight 5: Finished
  KkCtzC dtls TRACE: 06:55:19.400238 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF
  panic: EOF

  goroutine 36 [running]:
  github.com/pion/dtls/v2/examples/util.Check(0x652d00, 0xc000098100)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:65 +0x1b4
  github.com/pion/dtls/v2/examples/util.Chat.func1(0x654180, 0xc0000da580)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:31 +0x152
  created by github.com/pion/dtls/v2/examples/util.Chat
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:26 +0x64

  [6]    11792 exit 2     ./build/psk <<< 'exit'
  nTjWbB dtls TRACE: 06:55:19.404457 handshaker.go:151: [handshake:client] Flight 5: Preparing
  nTjWbB dtls TRACE: 06:55:19.404517 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  nTjWbB dtls TRACE: 06:55:19.404523 handshaker.go:151: [handshake:client] Flight 5: Sending
  nTjWbB dtls TRACE: 06:55:19.404528 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  nTjWbB dtls TRACE: 06:55:19.404535 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  nTjWbB dtls TRACE: 06:55:19.404588 handshaker.go:151: [handshake:client] Flight 5: Waiting
  nTjWbB dtls TRACE: 06:55:19.404956 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  nTjWbB dtls TRACE: 06:55:19.405102 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  nTjWbB dtls TRACE: 06:55:19.405111 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 145ms
  VkBrfh dtls TRACE: 06:55:19.446338 handshaker.go:151: [handshake:client] Flight 5: Finished
  VkBrfh dtls TRACE: 06:55:19.446362 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF
  panic: EOF

  goroutine 20 [running]:
  github.com/pion/dtls/v2/examples/util.Check(0x652d00, 0xc000012120)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:65 +0x1b4
  github.com/pion/dtls/v2/examples/util.Chat.func1(0x654180, 0xc00011e580)
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:31 +0x152
  created by github.com/pion/dtls/v2/examples/util.Chat
          /home/jdebruijn/withthegrid/projects/dtls/examples/util/util.go:26 +0x64

  [11]  + 11797 exit 2     ./build/psk <<< 'exit'
  Akl6hf dtls TRACE: 06:55:19.466718 handshaker.go:151: [handshake:client] Flight 5: Preparing
  Akl6hf dtls TRACE: 06:55:19.466782 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  Akl6hf dtls TRACE: 06:55:19.466789 handshaker.go:151: [handshake:client] Flight 5: Sending
  Akl6hf dtls TRACE: 06:55:19.466794 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  Akl6hf dtls TRACE: 06:55:19.466802 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  Akl6hf dtls TRACE: 06:55:19.466852 handshaker.go:151: [handshake:client] Flight 5: Waiting
  Akl6hf dtls TRACE: 06:55:19.467306 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  Akl6hf dtls TRACE: 06:55:19.467372 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  Akl6hf dtls TRACE: 06:55:19.467508 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 181ms
  Gh42hw dtls TRACE: 06:55:19.494083 handshaker.go:151: [handshake:client] Flight 5: Finished
  Gh42hw dtls TRACE: 06:55:19.494320 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF

  [9]  + 11795 done       ./build/psk <<< 'exit'
  IcPwE7 dtls TRACE: 06:55:19.528735 handshaker.go:151: [handshake:client] Flight 5: Finished
  IcPwE7 dtls TRACE: 06:55:19.528773 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF

  [2]    11788 done       ./build/psk <<< 'exit'
  WpckXj dtls TRACE: 06:55:19.529788 handshaker.go:151: [handshake:client] Flight 5: Preparing
  WpckXj dtls TRACE: 06:55:19.529852 handshaker.go:226: [handshake:client] -> changeCipherSpec (epoch: 1)
  WpckXj dtls TRACE: 06:55:19.529857 handshaker.go:151: [handshake:client] Flight 5: Sending
  WpckXj dtls TRACE: 06:55:19.529862 conn.go:372: [handshake:client] -> ClientKeyExchange (epoch: 0, seq: 2)
  WpckXj dtls TRACE: 06:55:19.529869 conn.go:372: [handshake:client] -> Finished (epoch: 1, seq: 3)
  WpckXj dtls TRACE: 06:55:19.529923 handshaker.go:151: [handshake:client] Flight 5: Waiting
  WpckXj dtls TRACE: 06:55:19.530175 conn.go:728: client: <- ChangeCipherSpec (epoch: 1)
  WpckXj dtls TRACE: 06:55:19.530208 handshaker.go:274: [handshake:client] Flight 5 -> Flight 5
  WpckXj dtls TRACE: 06:55:19.530216 handshaker.go:165: [handshake:client] Flight 5: Finished but sleeping for 86ms
  nTjWbB dtls TRACE: 06:55:19.550558 handshaker.go:151: [handshake:client] Flight 5: Finished
  nTjWbB dtls TRACE: 06:55:19.550610 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF

  [8]  + 11794 done       ./build/psk <<< 'exit'
  WpckXj dtls TRACE: 06:55:19.616321 handshaker.go:151: [handshake:client] Flight 5: Finished
  WpckXj dtls TRACE: 06:55:19.616354 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF

  [4]  - 11790 done       ./build/psk <<< 'exit'
  Akl6hf dtls TRACE: 06:55:19.648615 handshaker.go:151: [handshake:client] Flight 5: Finished
  Akl6hf dtls TRACE: 06:55:19.648638 conn.go:198: Handshake Completed
  Connected; type 'exit' to shutdown gracefully
  error: EOF

  [7]  + 11793 done       ./build/psk <<< 'exit'
  ```

  </details>

---

---

<h1 align="center">
  <br>
  Pion DTLS
  <br>
</h1>
<h4 align="center">A Go implementation of DTLS</h4>
<p align="center">
  <a href="https://pion.ly"><img src="https://img.shields.io/badge/pion-dtls-gray.svg?longCache=true&colorB=brightgreen" alt="Pion DTLS"></a>
  <a href="https://sourcegraph.com/github.com/pion/dtls"><img src="https://sourcegraph.com/github.com/pion/dtls/-/badge.svg" alt="Sourcegraph Widget"></a>
  <a href="https://pion.ly/slack"><img src="https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=brightgreen" alt="Slack Widget"></a>
  <br>
  <a href="https://travis-ci.org/pion/dtls"><img src="https://travis-ci.org/pion/dtls.svg?branch=master" alt="Build Status"></a>
  <a href="https://pkg.go.dev/github.com/pion/dtls"><img src="https://godoc.org/github.com/pion/dtls?status.svg" alt="GoDoc"></a>
  <a href="https://codecov.io/gh/pion/dtls"><img src="https://codecov.io/gh/pion/dtls/branch/master/graph/badge.svg" alt="Coverage Status"></a>
  <a href="https://goreportcard.com/report/github.com/pion/dtls"><img src="https://goreportcard.com/badge/github.com/pion/dtls" alt="Go Report Card"></a>
  <a href="https://www.codacy.com/app/Sean-Der/dtls"><img src="https://api.codacy.com/project/badge/Grade/18f4aec384894e6aac0b94effe51961d" alt="Codacy Badge"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
</p>
<br>

Native [DTLS 1.2][rfc6347] implementation in the Go programming language.

A long term goal is a professional security review, and maye inclusion in stdlib.

[rfc6347]: https://tools.ietf.org/html/rfc6347

### Goals/Progress
This will only be targeting DTLS 1.2, and the most modern/common cipher suites.
We would love contributes that fall under the 'Planned Features' and fixing any bugs!

#### Current features
* DTLS 1.2 Client/Server
* Key Exchange via ECDHE(curve25519, nistp256, nistp384) and PSK
* Packet loss and re-ordering is handled during handshaking
* Key export ([RFC 5705][rfc5705])
* Serialization and Resumption of sessions
* Extended Master Secret extension ([RFC 7627][rfc7627])

[rfc5705]: https://tools.ietf.org/html/rfc5705
[rfc7627]: https://tools.ietf.org/html/rfc7627

#### Supported ciphers

##### ECDHE
* TLS_ECDHE_ECDSA_WITH_AES_128_CCM ([RFC 6655][rfc6655])
* TLS_ECDHE_ECDSA_WITH_AES_128_CCM_8 ([RFC 6655][rfc6655])
* TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256 ([RFC 5289][rfc5289])
* TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256 ([RFC 5289][rfc5289])
* TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA ([RFC 8422][rfc8422])
* TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA ([RFC 8422][rfc8422])

##### PSK
* TLS_PSK_WITH_AES_128_CCM ([RFC 6655][rfc6655])
* TLS_PSK_WITH_AES_128_CCM_8 ([RFC 6655][rfc6655])
* TLS_PSK_WITH_AES_128_GCM_SHA256 ([RFC 5487][rfc5487])

[rfc5289]: https://tools.ietf.org/html/rfc5289
[rfc8422]: https://tools.ietf.org/html/rfc8422
[rfc6655]: https://tools.ietf.org/html/rfc6655
[rfc5487]: https://tools.ietf.org/html/rfc5487

#### Planned Features
* Chacha20Poly1305

#### Excluded Features
* DTLS 1.0
* Renegotiation
* Compression

### Using

This library needs at least Go 1.13, and you should have [Go modules
enabled](https://github.com/golang/go/wiki/Modules).

#### Pion DTLS
For a DTLS 1.2 Server that listens on 127.0.0.1:4444
```sh
go run examples/listen/selfsign/main.go
```

For a DTLS 1.2 Client that connects to 127.0.0.1:4444
```sh
go run examples/dial/selfsign/main.go
```

#### OpenSSL
Pion DTLS can connect to itself and OpenSSL.
```
  // Generate a certificate
  openssl ecparam -out key.pem -name prime256v1 -genkey
  openssl req -new -sha256 -key key.pem -out server.csr
  openssl x509 -req -sha256 -days 365 -in server.csr -signkey key.pem -out cert.pem

  // Use with examples/dial/selfsign/main.go
  openssl s_server -dtls1_2 -cert cert.pem -key key.pem -accept 4444

  // Use with examples/listen/selfsign/main.go
  openssl s_client -dtls1_2 -connect 127.0.0.1:4444 -debug -cert cert.pem -key key.pem
```

### Using with PSK
Pion DTLS also comes with examples that do key exchange via PSK


#### Pion DTLS
```sh
go run examples/listen/psk/main.go
```

```sh
go run examples/dial/psk/main.go
```

#### OpenSSL
```
  // Use with examples/dial/psk/main.go
  openssl s_server -dtls1_2 -accept 4444 -nocert -psk abc123 -cipher PSK-AES128-CCM8

  // Use with examples/listen/psk/main.go
  openssl s_client -dtls1_2 -connect 127.0.0.1:4444 -psk abc123 -cipher PSK-AES128-CCM8
```

### Contributing
Check out the **[contributing wiki](https://github.com/pion/webrtc/wiki/Contributing)** to join the group of amazing people making this project possible:

* [Sean DuBois](https://github.com/Sean-Der) - *Original Author*
* [Michiel De Backker](https://github.com/backkem) - *Public API*
* [Chris Hiszpanski](https://github.com/thinkski) - *Support Signature Algorithms Extension*
* [Iñigo Garcia Olaizola](https://github.com/igolaizola) - *Serialization & resumption, cert verification, E2E*
* [Daniele Sluijters](https://github.com/daenney) - *AES-CCM support*
* [Jin Lei](https://github.com/jinleileiking) - *Logging*
* [Hugo Arregui](https://github.com/hugoArregui)
* [Lander Noterman](https://github.com/LanderN)
* [Aleksandr Razumov](https://github.com/ernado) - *Fuzzing*
* [Ryan Gordon](https://github.com/ryangordon)
* [Stefan Tatschner](https://rumpelsepp.org/contact.html)
* [Hayden James](https://github.com/hjames9)
* [Jozef Kralik](https://github.com/jkralik)
* [Robert Eperjesi](https://github.com/epes)
* [Atsushi Watanabe](https://github.com/at-wat)
* [Julien Salleyron](https://github.com/juliens) - *Server Name Indication*
* [Jeroen de Bruijn](https://github.com/vidavidorra)
* [bjdgyc](https://github.com/bjdgyc)
* [Jeffrey Stoke (Jeff Ctor)](https://github.com/jeffreystoke) - *Fragmentbuffer Fix*
* [Frank Olbricht](https://github.com/folbricht)

### License
MIT License - see [LICENSE](LICENSE) for full text
