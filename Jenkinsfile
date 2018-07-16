node {
    checkout scm
        
    stage('Docker Build') {
        docker.build('dukfaar/recipebackend')
    }

    stage('Update Service') {
        sh 'docker service update --force recipebackend_recipebackend'
    }
}
