docker build -t istio-adapter-demo:1.0 -f Dockerfile .

docker save istio-adapter-demo:1.0 > /Users/hchoudhary/istio-adapter-demo_latest.tar

minikube ssh "docker load -i /Users/hchoudhary/istio-adapter-demo_latest.tar"

