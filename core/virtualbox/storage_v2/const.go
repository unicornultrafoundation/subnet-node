package storage

var userDataTemplate = `
#cloud-config
users:
  - name: %s
    plain_text_passwd: %s
    lock_passwd: false
    groups: sudo
    shell: /bin/bash

chpasswd:
  list: |
    %s:%s
  expire: false

ssh_pwauth: true
	`

var metaDataTemplate = `
instance-id: %s
local-hostname: %s
`
