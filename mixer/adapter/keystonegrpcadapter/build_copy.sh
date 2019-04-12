docker build -t istio/istio-keystone-adapter:1.0 -f Dockerfile .

docker save istio/istio-keystone-adapter:1.0 > /Users/hchoudhary/istio-keystone-grpcadapter_1.0.tar

minikube ssh "docker load -i /Users/hchoudhary/istio-keystone-grpcadapter_1.0.tar"
