{{/*
Expand the name of the chart.
*/}}
{{- define "stalkerr.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "stalkerr.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "stalkerr.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "stalkerr.labels" -}}
helm.sh/chart: {{ include "stalkerr.chart" . }}
{{ include "stalkerr.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "stalkerr.selectorLabels" -}}
app.kubernetes.io/name: {{ include "stalkerr.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "stalkerr.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "stalkerr.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Get the secret name
*/}}
{{- define "stalkerr.secretName" -}}
{{- if .Values.secrets.existingSecret }}
{{- .Values.secrets.existingSecret }}
{{- else }}
{{- include "stalkerr.fullname" . }}
{{- end }}
{{- end }}

{{/*
Get the ConfigMap name
*/}}
{{- define "stalkerr.configMapName" -}}
{{- include "stalkerr.fullname" . }}
{{- end }}

{{/*
Get the PostgreSQL service name
*/}}
{{- define "stalkerr.postgresql.serviceName" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" (include "stalkerr.fullname" .) }}
{{- else }}
{{- .Values.config.database.host }}
{{- end }}
{{- end }}

{{/*
Get the PostgreSQL port
*/}}
{{- define "stalkerr.postgresql.port" -}}
{{- if .Values.postgresql.enabled }}
{{- print "5432" }}
{{- else }}
{{- .Values.config.database.port | toString }}
{{- end }}
{{- end }}

{{/*
Get the PostgreSQL database name
*/}}
{{- define "stalkerr.postgresql.database" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.database }}
{{- else }}
{{- .Values.config.database.name }}
{{- end }}
{{- end }}

{{/*
Get the PostgreSQL username
*/}}
{{- define "stalkerr.postgresql.username" -}}
{{- if .Values.postgresql.enabled }}
{{- .Values.postgresql.auth.username }}
{{- else }}
{{- .Values.config.database.user }}
{{- end }}
{{- end }}

{{/*
Compose the DATABASE_URL environment variable
*/}}
{{- define "stalkerr.databaseUrl" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "postgres://%s:$(STALKERR_DATABASE_PASSWORD)@%s:%s/%s" 
    (include "stalkerr.postgresql.username" .) 
    (include "stalkerr.postgresql.serviceName" .) 
    (include "stalkerr.postgresql.port" .) 
    (include "stalkerr.postgresql.database" .) }}
{{- else }}
{{- printf "postgres://%s:$(STALKERR_DATABASE_PASSWORD)@%s:%s/%s?sslmode=%s" 
    .Values.config.database.user 
    .Values.config.database.host 
    (.Values.config.database.port | toString) 
    .Values.config.database.name 
    .Values.config.database.sslmode }}
{{- end }}
{{- end }}

{{/*
Get the image name
*/}}
{{- define "stalkerr.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.image.repository $tag }}
{{- end }}

{{/*
Get M3U PVC name
*/}}
{{- define "stalkerr.m3uPvcName" -}}
{{- if .Values.storage.m3u.existingClaim }}
{{- .Values.storage.m3u.existingClaim }}
{{- else }}
{{- printf "%s-m3u" (include "stalkerr.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Get media PVC name
*/}}
{{- define "stalkerr.mediaPvcName" -}}
{{- if .Values.storage.media.existingClaim }}
{{- .Values.storage.media.existingClaim }}
{{- else }}
{{- printf "%s-media" (include "stalkerr.fullname" .) }}
{{- end }}
{{- end }}
