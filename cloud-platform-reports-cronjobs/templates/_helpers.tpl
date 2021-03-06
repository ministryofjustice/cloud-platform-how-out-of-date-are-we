{{/*

  Defines the host to which JSON data should be POSTed,
  and the API key to authenticate POSTed data.

*/}}
{{- define "cloud-platform-reports-cronjobs.hoodaw-credentials" }}
- name: HOODAW_API_KEY
  valueFrom:
    secretKeyRef:
      name: {{ .Values.webApplication.apiSecretName }}
      key: {{ .Values.webApplication.apiSecretKey }}
- name: HOODAW_HOST
  value: {{ .Values.webApplication.hostname }}
{{- end }}

{{/*

  Add imagePullSecrets so that the cronjob uses credentials from the named
  docker-registry secret, rather than pulling images anonymously.

*/}}
{{- define "cloud-platform-reports-cronjobs.imagePullSecrets" }}
imagePullSecrets:
  - name: {{ .Values.dockerhubCredentials.secretName }}
{{- end }}

{{/*

  AWS credentials

  These are expected to already be in the namespace, in the `aws-creds` secret.

*/}}
{{- define "cloud-platform-reports-cronjobs.awsCredentials" }}
- name: AWS_REGION
  value: eu-west-2
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: aws-creds
      key: access-key-id
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: aws-creds
      key: secret-access-key
{{- end }}


{{/*

  Kubeconfig

  The S3 bucket and location of the cluster kubeconfig file

*/}}
{{- define "cloud-platform-reports-cronjobs.kubeconfigLocation" }}
- name: KUBECONFIG
  value: /tmp/kubeconfig
- name: KUBECONFIG_S3_BUCKET
  value: cloud-platform-concourse-kubeconfig
- name: KUBECONFIG_S3_KEY
  value: kubeconfig
{{- end }}
