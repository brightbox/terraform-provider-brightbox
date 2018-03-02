pipeline {
  options {
    disableConcurrentBuilds()
    buildDiscarder(logRotator(numToKeepStr: '5'))
  }
  triggers {
    cron('@weekly')
  }
  agent none
  stages {
    stage("Test") {
      agent {
        docker {
	  image 'golang'
	  label "docker"
	}
      }
      steps {
	sh 'go get -u github.com/jstemmer/go-junit-report'
        sh 'go get -t -v -d'
        sh 'go test -v | go-junit-report | tee report.xml'
	sh 'pwd'
      }
      post {
        always {
          junit "report.xml"
        }
      }
    }
  }
}
