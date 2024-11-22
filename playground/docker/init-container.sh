#!/bin/sh
while
     kaon-cli help &> /dev/null
     rc=$?; if [[ $rc == 0 ]]; then break; fi
do :;  done

balance=`kaon-cli getbalance`
if [ "${balance:0:1}" == "0" ]
then
    set -x
	kaon-cli generate 600 > /dev/null
	set -
fi

WALLETFILE=test-wallet
LOCKFILE=${KAON_DATADIR}/import-test-wallet.lock

if [ ! -e $LOCKFILE ]; then
  while
       kaon-cli getaddressesbyaccount "" &> /dev/null
       rc=$?; if [[ $rc != 0 ]]; then continue; fi

       set -x
       
       kaon-cli importprivkey "YoBkvJXYcxihY3qYXZuh3ndN14hWmAPUU3qhFZw1Zbt9g85FjMCd" # addr=auASFMxv45WgjCW6wkpDuHWjxXhzNA9mjP

       # MM private key for this addr is cbc9b23fc49066bbe19e599364035b9e8d11bb51e0f1fb56b14f50564bfd15e9
       # there is no private key for this addr in KAON
       solar prefund ar2SzdHghSgeacypPn7zfDe3qfKAEwimus 500
       # KAON private key for this addr is YoBkvJXYcxihY3qYXZuh3ndN14hWmAPUU3qhFZw1Zbt9g85FjMCd
       # there is no private key for this addr in Eth format
       solar prefund auASFMxv45WgjCW6wkpDuHWjxXhzNA9mjP 500
       touch $LOCKFILE

       set -
       break
  do :;  done
fi
