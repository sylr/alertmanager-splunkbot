Forwarding alerts sent by prometheus alertmaneger to splunk in order to understand the logic ...

Alertmanager splunkbot k8s container:

```yaml
      - name: prometheus-alertmanager-splunkbot
        image: sylr/alertmanager-splunkbot:v0.0.6
        args:
        - -v
        - --insecure
        env:
        - name: SPLUNKBOT_LISTENING_ADDRESS
          value: 127.0.0.1
        - name: SPLUNKBOT_LISTENING_PORT
          value: "44553"
        - name: SPLUNKBOT_SPLUNK_URL
          value: https://10.101.0.46/services/collector/event/1.0
        - name: SPLUNKBOT_SPLUNK_TOKEN
          valueFrom:
            secretKeyRef:
              key: splunk-token
              name: splunkbot-secrets
        imagePullPolicy: IfNotPresent
        resources: {}
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
```

Alertmanager config:

```yaml
receivers:
  webhook_configs:
  - send_resolved: true
    url: http://127.0.0.1:44553
```
