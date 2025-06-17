#!/bin/bash

# Function to create a test UAV
create_test_uav() {
    local name=$1
    local battery=$2
    local motor_consumption=$3
    local ce_consumption=$4

    cat <<EOF | kubectl apply -f -
apiVersion: fanet.restart.eu/v1alpha1
kind: UAV
metadata:
  name: ${name}
spec:
  nodeName: "kind-worker"
  batteryLevel: ${battery}
  flightStatus: "In-flight"
  energyConsumption:
    motorIdleConsumption: "${motor_consumption}"
    ceIdleConsumption: "${ce_consumption}"
  ceCapacity:
    maxCapacity: 5
    deployedVFs: 0
  networkMetrics:
    interUAVDelay: 50
EOF
}

# Function to deploy the simulator
deploy_simulator() {
    kubectl apply -f config/simulator/simulator.yaml
}

# Function to watch UAV status
watch_uav_status() {
    kubectl get uav -w
}

# Function to clean up
cleanup() {
    kubectl delete -f config/simulator/simulator.yaml
    kubectl delete uav --all
}

# Main script
case "$1" in
    "deploy")
        deploy_simulator
        ;;
    "create")
        if [ "$#" -ne 5 ]; then
            echo "Usage: $0 create <name> <battery> <motor_consumption> <ce_consumption>"
            exit 1
        fi
        create_test_uav "$2" "$3" "$4" "$5"
        ;;
    "watch")
        watch_uav_status
        ;;
    "cleanup")
        cleanup
        ;;
    *)
        echo "Usage: $0 {deploy|create|watch|cleanup}"
        echo "  deploy: Deploy the UAV simulator"
        echo "  create <name> <battery> <motor_consumption> <ce_consumption>: Create a test UAV"
        echo "  watch: Watch UAV status changes"
        echo "  cleanup: Remove simulator and all UAVs"
        exit 1
        ;;
esac 