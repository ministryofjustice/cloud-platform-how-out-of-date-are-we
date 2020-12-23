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
