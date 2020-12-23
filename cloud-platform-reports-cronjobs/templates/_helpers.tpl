{{/*

  Defines the host to which JSON data should be POSTed,
  and the API key to authenticate POSTed data.

*/}}
{{- define "cloud-platform-reports-cronjobs.hoodaw-credentials" }}
            - name: HOODAW_API_KEY
              valueFrom:
                secretKeyRef:
                  name: hoodaw-creds
                  key: api_key
            - name: HOODAW_HOST
              value: "https://{{ .Values.ingress.hostname }}"
{{- end }}

{{/*

  Add imagePullSecrets so that the cronjob uses credentials from the named
  docker-registry secret, rather than pulling images anonymously.

*/}}
{{- define "cloud-platform-reports-cronjobs.imagePullSecrets" }}
          imagePullSecrets:
            - name: {{ .Values.dockerhubCredentials.secretName }}
{{- end }}
