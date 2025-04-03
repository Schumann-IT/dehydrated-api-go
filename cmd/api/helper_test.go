package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func setupAzureDnsHook(baseDir string, t *testing.T) string {
	// Create azure-dns hook
	hookFile := filepath.Join(baseDir, "hook.sh")
	hookData := []byte(fmt.Sprintf(`
#!/usr/bin/env bash

#
# How to deploy a DNS challenge using Azure
#

# Debug Logging level
DEBUG=4

# Azure Tenant specific configuration settings
#   You should create an SPN in Azure first and authorize it to make changes to Azure DNS
#       REF: https://azure.microsoft.com/en-us/documentation/articles/resource-group-create-service-principal-portal/
AZURE_TENANT="%s"
AZURE_CLIENT_ID="%s"
AZURE_CLIENT_SECRET="%s"
SUBSCRIPTION="%s"
RESOURCE_GROUP="%s"
DNS_ZONE="%s"

# Supporting functions
function log {
    if [ $DEBUG -ge $2 ]; then
        echo "$1" > /dev/tty
    fi
}
function login_azure {
    # Azure DNS Connection Variables
    # You should create an SPN in Azure first and authorize it to make changes to Azure DNS
    #  REF: https://azure.microsoft.com/en-us/documentation/articles/resource-group-create-service-principal-portal/
    az login -u ${AZURE_CLIENT_ID} -p ${AZURE_CLIENT_SECRET} --tenant ${AZURE_TENANT} --service-principal > /dev/null
}
function parseSubDomain {
    log "  Parse SubDomain" 4

    FQDN="$1"
    log "    FQDN: '${FQDN}'" 4

    DOMAIN=${DNS_ZONE}
    log "    DOMAIN: '${DOMAIN}'" 4

    SUBDOMAIN=$(sed -E "s/(.*)\.${DNS_ZONE//./\.}/\1/" <<< "${FQDN}")
    log "    SUBDOMAIN: '${SUBDOMAIN}'" 4

    echo "${SUBDOMAIN}"
}
function buildDnsKey {
    log "  Build DNS Key" 4

    FQDN="$1"
    log "    FQDN: '${FQDN}'" 4

    SUBDOMAIN=$(parseSubDomain ${FQDN})
    log "    SUBDOMAIN: ${SUBDOMAIN}" 4

    CHALLENGE_KEY="_acme-challenge.${SUBDOMAIN}"
    log "    KEY: '${CHALLENGE_KEY}'" 4

    echo "${CHALLENGE_KEY}"
}


# Logging the header
log "Azure Hook Script - LetsEncrypt" 4


# Execute the specified phase
PHASE="$1"
log "" 1
log "  Phase: '${PHASE}'" 1
#log "    Arguments: ${1} | ${2} | ${3} | ${4} | ${5} | ${6} | ${7} | ${8} | ${9} | ${10}" 1
case ${PHASE} in
    'deploy_challenge')
        login_azure

        # Arguments: PHASE; DOMAIN; TOKEN_FILENAME; TOKEN_VALUE
        FQDN="$2"
        TOKEN_VALUE="$4"
        SUBDOMAIN=$(parseSubDomain ${FQDN})
        CHALLENGE_KEY=$(buildDnsKey ${FQDN})

        # Commands
        log "" 4
        log "    Running azure cli commands" 4

        respCreate=$(az network dns record-set txt create --subscription ${SUBSCRIPTION} -g ${RESOURCE_GROUP} -z ${DNS_ZONE} -n ${CHALLENGE_KEY} --output json)
        log "      Create: '$respCreate'" 4

        respAddRec=$(az network dns record-set txt add-record --subscription ${SUBSCRIPTION} -g ${RESOURCE_GROUP} -z ${DNS_ZONE} -n ${CHALLENGE_KEY} -v ${TOKEN_VALUE} --output json)
        log "      AddRec: '$respAddRec'" 4
        ;;

    "clean_challenge")
        login_azure

        # Arguments: PHASE; DOMAIN; TOKEN_FILENAME; TOKEN_VALUE
        FQDN="$2"
        TOKEN_VALUE="$4"
        SUBDOMAIN=$(parseSubDomain ${FQDN})
        CHALLENGE_KEY=$(buildDnsKey ${FQDN})

        # Commands
        log "" 4
        log "    Running azure cli commands" 4

        respDel=$(az network dns record-set txt delete --subscription ${SUBSCRIPTION} -g ${RESOURCE_GROUP} -z ${DNS_ZONE} -n ${CHALLENGE_KEY} -y --output json)
        log "      Delete: '$respDel'" 4
        ;;

    "deploy_cert")
        # Parameters:
        # - PHASE           - the phase being executed
        # - DOMAIN          - the domain name (CN or subject alternative name) being validated.
        # - KEY_PATH        - the path to the certificate's private key file
        # - CERT_PATH       - the path to the certificate file
        # - FULL_CHAIN_PATH - the path to the full chain file
        # - CHAIN_PATH      - the path to the chain file
        # - TIMESTAMP       - the timestamp of the deployment

        # do nothing for now
        ;;

    "unchanged_cert")
        # Parameters:
        # - PHASE           - the phase being executed
        # - DOMAIN          - the domain name (CN or subject alternative name) being validated.
        # - KEY_PATH        - the path to the certificate's private key file
        # - CERT_PATH       - the path to the certificate file
        # - FULL_CHAIN_PATH - the path to the full chain file
        # - CHAIN_PATH      - the path to the chain file

        # do nothing for now
        ;;

    *)
        #log "Unknown hook '${PHASE}'" 1
        exit 0
        ;;
esac

exit 0

`, os.Getenv("AZURE_TENANT"), os.Getenv("AZURE_CLIENT_ID"), os.Getenv("AZURE_CLIENT_SECRET"), os.Getenv("SUBSCRIPTION"), os.Getenv("RESOURCE_GROUP"), os.Getenv("DNS_ZONE")))
	if err := os.WriteFile(hookFile, hookData, 0755); err != nil {
		t.Fatalf("Failed to write hook file: %v", err)
	}

	return hookFile
}

func setupDehydratedConfig(baseDir, hookScript, algo string, t *testing.T) {
	// Create test dehydrated config file
	dehydratedConfigFile := filepath.Join(baseDir, "config")
	dehydratedConfigData := []byte(fmt.Sprintf(`
CHALLENGETYPE="dns-01"
HOOK="%s"
KEY_ALGO=%s
`, hookScript, algo))
	if err := os.WriteFile(dehydratedConfigFile, dehydratedConfigData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
}

func setupDomains(baseDir string, t *testing.T) {
	// Create domains config file
	domainsFile := filepath.Join(baseDir, "domains.txt")
	domainsData := []byte(`
foo.hq.schumann-it.com
*.whatever.hq.schumann-it.com > whatever
`)
	if err := os.WriteFile(domainsFile, domainsData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
}
