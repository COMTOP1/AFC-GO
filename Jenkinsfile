String registryEndpoint = 'registry.registry-01.dmxhbjIwbG9jYWwK.bswdi.co.uk'

def branch = env.BRANCH_NAME.replaceAll("/", "_")
def image
String imageName = "bswdi/afc-go:${branch}-${env.BUILD_ID}"

pipeline {
  agent {
    label 'docker'
  }

  environment {
    DOCKER_BUILDKIT = '1'
  }

  stages {
    stage('Build image') {
      steps {
        script {
          def GIT_COMMIT_HASH = sh (script: "git log -n 1 --pretty=format:'%H' | head -c 7", returnStdout: true)
          docker.withRegistry('https://' + registryEndpoint) {
            image = docker.build(imageName, "--build-arg AFC_VERSION_ARG=${env.BRANCH_NAME}-${env.BUILD_ID} --build-arg AFC_COMMIT_ARG=${GIT_COMMIT_HASH} --no-cache .")
          }
        }
      }
    }

    stage('Push image to registry') {
      steps {
        script {
          docker.withRegistry('https://' + registryEndpoint) {
            image.push()
            if (env.BRANCH_IS_PRIMARY) {
              image.push('latest')
            }
            if (env.CHANGE_ID) {
              image.push("pr-${env.CHANGE_ID}")
            }
          }
        }
      }
    }

    stage('Deploy') {
      stages {
        stage('Preview') {
          when {
            changeRequest target: 'main'
          }
          stages {
            stage('Cleanup previews') {
              agent {
                label 'nomad'
              }
              steps {
                deployPreview action: 'cleanup'
                deployPreview action: 'cleanupMerge'
              }
            }
            stage('Preview') {
              steps {
                deployPreview action: 'deploy', job: 'afc-go/preview', jobName: 'afc-go-preview', urlSuffix: 'preview.afcaldermaston.co.uk'
              }
            }
          }
        }

        stage('Development') {
          when {
            expression { env.BRANCH_IS_PRIMARY }
          }
          stages {
            stage('Deploy to development') {
              steps {
                build job: 'Deploy Nomad Job', parameters: [
                  string(name: 'JOB_FILE', value: 'afc-go-dev.nomad'),
                  text(name: 'TAG_REPLACEMENTS', value: "${registryEndpoint}/${imageName}")
                ], wait: true
              }
            }

            stage('Post-deploy cleanup & migrate') {
              agent {
                label 'nomad'
              }
              steps {
                checkout scm
                deployPreview action: 'cleanup'
                deployPreview action: 'cleanupMerge'
              }
            }
          }
        }

        stage('Production') {
          when {
            // Checking if it is semantic version release.
            expression { return env.TAG_NAME ==~ /v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)/ }
          }
          steps {
            build(job: 'Deploy Nomad Job', parameters: [
              string(name: 'JOB_FILE', value: 'afc-go-prod.nomad'),
              text(name: 'TAG_REPLACEMENTS', value: "${registryEndpoint}/${imageName}")
            ])
          }
        }
      }
    }
  }
}
