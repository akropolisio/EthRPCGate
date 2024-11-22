#!/bin/sh
repeat_until_success () {
    echo Running command - "$@"
    i=0
    until $@
    do
        echo Command failed with exit code - $?
        if [ $i -gt 10 ]; then
            echo Giving up running command - "$@"
            return
        fi
        echo Sleeping $i seconds
        sleep $i
        echo Retrying
        i=`expr $i + 1`
    done
    echo Command finished successfully
}

#import private keys and then prefund them
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd importprivkey "YoBkvJXYcxihY3qYXZuh3ndN14hWmAPUU3qhFZw1Zbt9g85FjMCd" address2 # addr=auASFMxv45WgjCW6wkpDuHWjxXhzNA9mjP hdkeypath=m/88'/0'/2'
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd importprivkey "Yn3MpyMQzQR7osXkNyQ3DVNP6kwZGpmXhXzuuPGXZwyouhxmWxh7" address3 # addr=awQb8vf21idkFoZiYPA4hWgtuPyko2qUaR
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd importprivkey "YhrMRdEunTgoz4vs9RYc7Ui5yAYp2UmBeZsVfWHjifzVWeDQGVt7" address4 # addr=ar7PkgNdY1HkDtUo3D4GTsYrcqoHBJygNQ
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd importprivkey "YiaCnmCToCPUGgSrLxEiGVa297WiFGc5jrsUpy68BJiCe7Cym716" address5 # addr=b7CSynDNwb2LQcCWXs8Qn79LUkgMdsK61S
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd importprivkey "YifXHT9WJeP53jor88a6FYgxj3CDFjtkbYGb6tRHJK3UhdLye6m2" address6 # addr=ayHXgXugbHDDR8cBjX2ZVLfkGE78QTeW2Z

echo Finished importing accounts
echo Seeding accounts
echo Seeding ucavTSEVe31NLdXyfq925GzGp8yN5QnS6a
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd generatetoaddress 2 ucavTSEVe31NLdXyfq925GzGp8yN5QnS6a
kaon-cli -rpcuser=kaon -rpcpassword=testpasswd generatetoaddress 2 ua1W7VnwtJPoFDoNjxxGdDHBtsRKDpjW8c
kaon-cli -rpcuser=kaon -rpcpassword=testpasswd generatetoaddress 1000 uaofg5zZVyvPmWgGL6YdVAyRTKWd3MjZ4A

# address1
# hex addr 0x1CE507204a6fC8fd6aA7e54D1481d30ACB0Dbead (ar2SzdHghSgeacypPn7zfDe3qfKAEwimus) has only MM private key: cbc9b23fc49066bbe19e599364035b9e8d11bb51e0f1fb56b14f50564bfd15e9
# there is no private key for this addr in KAON format
echo Seeding ar2SzdHghSgeacypPn7zfDe3qfKAEwimus
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd generatetoaddress 1000 ar2SzdHghSgeacypPn7zfDe3qfKAEwimus

# address2
# hex addr 0x3f501c368cb9ddb5f27ed72ac0d602724adfa175 (auASFMxv45WgjCW6wkpDuHWjxXhzNA9mjP) has only KAON private key: YoBkvJXYcxihY3qYXZuh3ndN14hWmAPUU3qhFZw1Zbt9g85FjMCd
# there is no private key for this addr in Eth format
echo Seeding auASFMxv45WgjCW6wkpDuHWjxXhzNA9mjP
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd generatetoaddress 1000 auASFMxv45WgjCW6wkpDuHWjxXhzNA9mjP
# address3
echo Seeding awQb8vf21idkFoZiYPA4hWgtuPyko2qUaR
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd generatetoaddress 500 awQb8vf21idkFoZiYPA4hWgtuPyko2qUaR
# address4
echo Seeding ar7PkgNdY1HkDtUo3D4GTsYrcqoHBJygNQ
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd generatetoaddress 250 ar7PkgNdY1HkDtUo3D4GTsYrcqoHBJygNQ
# address5
echo Seeding b7CSynDNwb2LQcCWXs8Qn79LUkgMdsK61S
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd generatetoaddress 100 b7CSynDNwb2LQcCWXs8Qn79LUkgMdsK61S
# address6
echo Seeding ayHXgXugbHDDR8cBjX2ZVLfkGE78QTeW2Z
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd generatetoaddress 100 ayHXgXugbHDDR8cBjX2ZVLfkGE78QTeW2Z
# playground pet shop dapp
# PK KAON: Yjf1pkLANZcMQd81HyyqV4BHBMwMLoqZzZ9W4YfGxtqEbinL5kQq
echo Seeding 0x270e4f191c2f13cfaea6c35edfe4020b433632d6
repeat_until_success kaon-cli -rpcuser=kaon -rpcpassword=testpasswd generatetoaddress 2 arxBCorh4mfs32wDvVpCx3fBeRjnTgMKNV
echo Finished importing and seeding accounts
