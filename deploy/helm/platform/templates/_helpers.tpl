{{- define "platform.name" -}}
{{- default .Chart.Name .Values.global.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "platform.fullname" -}}
{{- $name := include "platform.name" . -}}
{{- if .Values.global.fullnameOverride -}}
{{- .Values.global.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "platform.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{- define "platform.namespace" -}}
{{- default .Release.Namespace .Values.global.namespace.name -}}
{{- end -}}

{{- define "platform.componentFullname" -}}
{{- printf "%s-%s" (include "platform.fullname" .root) .service.component | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "platform.baseLabels" -}}
helm.sh/chart: {{ include "platform.chart" .root }}
app.kubernetes.io/name: {{ include "platform.name" .root }}
app.kubernetes.io/instance: {{ .root.Release.Name }}
app.kubernetes.io/version: {{ .root.Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .root.Release.Service }}
app.kubernetes.io/part-of: {{ include "platform.fullname" .root }}
{{- with .root.Values.global.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end -}}

{{- define "platform.labels" -}}
{{ include "platform.baseLabels" (dict "root" .root) }}
app.kubernetes.io/component: {{ .service.component }}
{{- end -}}

{{- define "platform.selectorLabels" -}}
app.kubernetes.io/name: {{ include "platform.name" .root }}
app.kubernetes.io/instance: {{ .root.Release.Name }}
app.kubernetes.io/component: {{ .service.component }}
{{- end -}}

{{- define "platform.serviceAccountName" -}}
{{- if .service.serviceAccount.name -}}
{{- .service.serviceAccount.name -}}
{{- else if .service.serviceAccount.create -}}
{{- include "platform.componentFullname" . -}}
{{- else -}}
default
{{- end -}}
{{- end -}}
