
DEPLOY_NAME="gpu-pod"

function check() {
    local target=$1
    local is_match="false"

    for i in $(seq 500)
    do
        AVAILABLE_REPLICA=$(kubectl get deploy ${DEPLOY_NAME} | awk '{print $4}' | tail -n 1)
        NODES=$(kubectl get nodes -l visenze.component=search,visenze.gpu=true -o json | jq '.items | length')
        if [ ${AVAILABLE_REPLICA} -eq ${target} ] && [ ${NODES} -eq ${target} ]
        then
            is_match="true"
            break
        fi
        sleep 5
    done
    echo ${is_match}
}

function scale() {
    kubectl scale --replicas=$1 deployment/${DEPLOY_NAME}
}

function scale_and_check() {
    local target=$1
    echo "Scaling to ${target}, checking..."
    scale ${target} 
    if [ $(check ${target}) = "false" ]
    then
        echo "Scaling to ${target} doesn't work"
        exit 1
    else
        echo "Scaling to ${target} succeed"
    fi
}

function update_resource_limit() {
    cp gpu-deploy-tmpl.yaml gpu-deploy-tmp.yaml
    local append_txt=""
    case $1 in
    gpu_num)
      append_txt="              nvidia.com/gpu: 1"    
      ;;

    gpu_memory)
      append_txt="              visenze.com/nvidia-gpu-memory: 8988051968"
      ;;

    mps_context)
      append_txt="              visenze.com/nvidia-mps-context: 18"
      ;;

    # *)
      # STATEMENTS
      # ;;
    esac
    echo "$append_txt" >> gpu-deploy-tmp.yaml
    kubectl apply -f gpu-deploy-tmp.yaml
}

function test_with_resource() {
    echo "check the resource $1"
    update_resource_limit $1
    scale_and_check 1
    scale_and_check 2
    scale_and_check 1
    scale_and_check 0
}


test_with_resource "gpu_num"
test_with_resource "gpu_memory"
test_with_resource "mps_context"
