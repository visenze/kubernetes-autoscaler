// throttle concurrent build
properties([[$class: 'ThrottleJobProperty', categories: [], limitOneJobWithMatchingParams: false, maxConcurrentPerNode: 1, maxConcurrentTotal: 1, paramsToUseForLimit: '', throttleEnabled: true, throttleOption: 'project']])

library(identifier: "visenze-lib@${params['VISENZE_LIB_BRANCH'] ?: 'master'}", changelog: false)

pipeline {
  agent {
    label "build-arm64"
  }

  options {
    ansiColor('xterm')
  }

  tools {
    go 'go1.15'
  }

  stages {
    stage('Checkout') {
      steps {
        checkout([
            $class: 'GitSCM',
            branches: scm.branches,
            doGenerateSubmoduleConfigurations: scm.doGenerateSubmoduleConfigurations,
            extensions: [
                [
                    $class: 'CloneOption',
                    noTags: true,
                    reference: '',
                    timeout: 60
                ],
                [
                    $class: 'SubmoduleOption',
                    disableSubmodules: false,
                    parentCredentials: true,
                    recursiveSubmodules: true,
                    trackingSubmodules: true,
                    reference: '',
                    timeout: 60
                ],
                [$class: 'CleanBeforeCheckout']
            ],
            userRemoteConfigs: scm.userRemoteConfigs
        ])
      }
    }

    stage('Test') {
      when {
        expression {
          return canRun()
        }
      }
      steps {
        script {
          dir('cluster-autoscaler') {
            sh('make test-in-docker')
          }
        }
      }
    }

    stage('Docker Build&Push') {
      when {
        expression {
          return canRun()
        }
      }
      steps {
        script {
          dir('cluster-autoscaler') {
            def version = sh(script: "grep ClusterAutoscalerVersion version/version.go",
                returnStdout: true).split('"')[-2]
            docker.withRegistry('', 'docker-hub-credential') {
              sh("docker buildx create --use")
              sh("docker buildx build -t visenze/cluster-autoscaler:${version} --push --platform linux/arm64,linux/amd64 .")
            }

            build(job: 'devops_docker_registry_copy_image', parameters: [
                string(name: 'REPOSITORY', value: "visenze/cluster-autoscaler"),
                string(name: 'DOCKER_TAG', value: version),
                string(name: 'TIMEOUT', value: "30"),
                string(name: 'SOURCE_DOCKER_REGISTRY_CREDENTIAL', value: "docker-hub-credential"),
                string(name: 'DEST_DOCKER_REGISTRY', value: "https://741813507711.dkr.ecr.cn-northwest-1.amazonaws.com.cn"),
                string(name: 'DEST_DOCKER_REGISTRY_CREDENTIAL', value: "ecr:cn-northwest-1:aws-cn-jenkins"),
            ])
          }
        }
      }
    }
  }
}

def canRun() {
  return env.BRANCH_NAME.startsWith('release-') || env.BRANCH_NAME == 'master'
}