#!/bin/bash

# Which public key algorithm should be used? Supported: rsa, prime256v1 and secp384r1
KEY_ALGO=prime256v1

KEYSIZE="4096"
PRIVATE_KEY_RENEW="yes"

#USE BELOW CA FOR TESTING OTHERWISE YOU MIGHT GET BANNED FROM LE https://community.letsencrypt.org/t/rate-limits-for-lets-encrypt/6769
#CA="https://acme-staging-v02.api.letsencrypt.org/directory"

# Minimum days before expiration to automatically renew certificate (default: 30)
# 90 Tage ist ein Cert gültig
RENEW_DAYS="60"

#HOOK="/home/le-user/dns-challenge/"
HOOK_CHAIN="no"

WELLKNOWN="/home/le-user/dns-challenge/www-temp"
# E-mail to use during the registration (default: <unset>)
CONTACT_EMAIL="webmaster@hansemerkur.de"

#Challenge Type
#Change to dns-01 for DNS challenge
CHALLENGETYPE="dns-01"

# OCSP Stapeling
# Option to add CSR-flag indicating OCSP stapling to be mandatory (default: no)
# OCSP_MUST_STAPLE="yes"
# Fetch OCSP responses (default: no)
#OCSP_FETCH="no"
# OCSP refresh interval (default: 5 days)
#OCSP_DAYS=5
