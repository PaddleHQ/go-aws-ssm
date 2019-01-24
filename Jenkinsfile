pipeline {
    agent any
    environment {
        CI = "true"
        UNIQUE_BUILD_ID = "${GIT_COMMIT}".substring(0,7)
    }
    stages {
        stage('Build') {
            steps {
                sh('make build')
            }
        }
        stage('Static analysis') {
            steps {
                sh('make lint vet')
            }
         }
        stage('Unit tests') {
            steps {
                sh('make unit-test')
            }
        }

    }
    post {
        failure {
            script {
                sh('make docker-dump-logs')
            }
        }
        cleanup {
            script {
                cleanWs() /* clean up our workspace */
            }
        }
    }
}
