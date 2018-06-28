pipeline {
  options {
    disableConcurrentBuilds()
    buildDiscarder(logRotator(numToKeepStr: '5'))
  }
  triggers {
    cron('@weekly')
  }
  agent {
    docker {
      image 'golang'
      label "docker"
      args "-v /tmp:/.cache"
    }
  }
  stages {
    stage("Prepare dependencies") {
      steps {
	sh 'go get -u github.com/jstemmer/go-junit-report'
        sh 'go get -t -v -d'
      }
    }
    stage("Test") {
      steps {
        sh 'go test -v | go-junit-report | tee report.xml'
      }
      post {
        always {
          junit "report.xml"
        }
      }
    }
  }
}
