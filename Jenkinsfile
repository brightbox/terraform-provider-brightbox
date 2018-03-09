pipeline {
  environment {
    BRIGHTBOX_ORBIT_URL = credentials('c0c95c56-b4e9-45c2-85c5-bf292aef7301')
    BUILDER = credentials('ced8ecd0-dbe4-4bf6-bdde-101a661565f5')
    BRIGHTBOX_CLIENT="${BUILDER_USR}"
    BRIGHTBOX_CLIENT_SECRET="${BUILDER_PSW}"
    BRIGHTBOX_API_URL = credentials('ab3d6198-ad49-4e27-9542-d69cc6a05cc5')
    GITHUB = credentials('03dee9c7-c93d-4ae0-b45c-3d2cb7b672c4')
    GITHUB_TOKEN = "${GITHUB_PSW}"
  }
  options {
    disableConcurrentBuilds()
    buildDiscarder(logRotator(numToKeepStr: '5'))
  }
  agent {
    docker {
      image 'golang'
      label 'docker'
      args '-v /tmp:/.cache'
    }
  }
  stages {
    stage("Vet") {
      steps {
        sh """
	target=\$(git ls-remote --get-url)
	target="\${target#https://}"
	target="/go/src/\${target%.git}"
	mkdir -p "\$(dirname \${target})"
        cp -a "$WORKSPACE" "\${target}"
	cd "\${target}"
	make vet
	"""
      }
    }
    stage("Acceptance Tests") {
      steps {
	sh """
	target=\$(git ls-remote --get-url)
	target="\${target#https://}"
	target="/go/src/\${target%.git}"
	cd "\${target}"
	make testaccjunit
	cp report.xml $WORKSPACE
	"""
      }
      post {
        always {
          junit "report.xml"
        }
      }
    }
    stage("Snapshot Build") {
      when {
        not { branch 'master' }
      }
      steps {
	sh """
	target=\$(git ls-remote --get-url)
	target="\${target#https://}"
	target="/go/src/\${target%.git}"
	cd "\${target}"
	RELEASEARGS="--snapshot" make release
	"""
      }
    }
    stage("Release Build") {
      when {
        branch 'master'
      }
      steps {
	sh """
	target=\$(git ls-remote --get-url)
	target="\${target#https://}"
	target="/go/src/\${target%.git}"
	cd "\${target}"
	make release
	"""
      }
    }
  }
}
