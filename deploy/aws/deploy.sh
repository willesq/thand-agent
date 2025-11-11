# You will need to first build and push the Docker image to ECR.
# Replace the ImageUri parameter value with your ECR image URI.

aws cloudformation deploy \
  --template-file cloudformation.yaml \
  --stack-name thand-agent \
  --parameter-overrides \
    ImageUri=12345.dkr.ecr.us-east-1.amazonaws.com/thand-io/agent:latest \
    ProvidersConfig="$(cat providers.yaml)" \
    RolesConfig="$(cat roles.yaml)" \
    WorkflowsConfig="$(cat workflows.yaml)" \
  --capabilities CAPABILITY_NAMED_IAM
