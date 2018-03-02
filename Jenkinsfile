pipeline {
  environment {
    ORBIT = credentials('8a05c5c4-d7fe-481d-b098-993503b80628')
    ORBIT_URL = credentials('c0c95c56-b4e9-45c2-85c5-bf292aef7301')
    ORBIT_USER = "${ORBIT_USR}"
    ORBIT_KEY = "${ORBIT_PSW}"
    BUILDER = credentials('ced8ecd0-dbe4-4bf6-bdde-101a661565f5')
    BRIGHTBOX_CLIENT="${BUILDER_USR}"
    BRIGHTBOX_CLIENT_SECRET="${BUILDER_PSW}"
    BRIGHTBOX_API_URL = credentials('ab3d6198-ad49-4e27-9542-d69cc6a05cc5')
    IMAGE_ACCOUNT = credentials('6fbcb860-104a-4167-a150-b9edec2c15f0')
    TF_VAR_distributions = '["artful", "bionic"]'
  }
  options {
    disableConcurrentBuilds()
    buildDiscarder(logRotator(numToKeepStr: '5'))
  }
  agent {
    docker {
      image 'golang'
      label 'docker'
    }
  }
  stages {
    stage("Vet") {
      steps {
        sh 'make vet'
      }
    }
    stage("Acceptance Tests") {
      steps {
	sh 'make testaccjunit'
      }
      post {
        always {
          junit "report.xml"
        }
      }
    }
  }
}
      


