go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

go tool pprof -inuse_space http://127.0.0.1:6060/debug/pprof/heap




go tool pprof -http=:54321 C:\Users\Colin\pprof\pprof.chat.exe.samples.cpu.001.pb.gz