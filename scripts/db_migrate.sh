#!/usr/bin/env bash
file=$1
packagePrefix=$2
goPackagePrefix=$3
#cat $file
rm -rf db_out
mkdir -p db_out/entity
mkdir -p db_out/proto_entity
table_list=($(grep CREATE ${file} | cut -d '`' -f 2))
function convert(){
    echo ${@}
}
for table in ${table_list[@]}; do
    echo "proc ==== ${table}===="
    tableUpper=`echo "${table}" | perl -pe 's/(^|_)./uc($&)/ge;s/_//g'`
    tableCmt=$(ssed -n '/CREATE.*'${table}'`/,/ENGINE=/p' ${file} | grep "ENGINE" | grep COMMENT | sed -e 's/.*COMMENT.*'\''\(.*\)'\''.*/\1/g')

    echo ${tableCmt}
    echo ${tableUpper}
    # 0.generate entity.go
    ssed -n '/CREATE.*'${table}'`/,/ENGINE=/p' ${file} | \
    grep -v CREATE | grep -v ') ENGINE' | grep -v KEY | \
    awk -F " " '
            function convertType(ttype){
                if ( match(type,/bigint/) ){
                    return "int64"
                }
                if ( match(type,/tinyint|enum/) ){
                    return "int32"
                }
                if ( match(type,/int/) ){
                    return "int64"
                }
                if ( match(type,/datetime|timestamp/) ){
                    return "int64"
                }
                if ( match(type,/varchar|text|char|blob|date/) ){
                    return "string"
                }

                if ( match(type,/decimal|double/) ){
                    return "float64"
                }

                return ttype
            }
            BEGIN{
                idx=0
                print "package entity"
                print "//'${tableCmt}'"
                print "type '${tableUpper}' struct {"
            }{

            gsub(/`/,"",$1)
            field=$1
            type=$2
            gsub(/.*COMMENT/,"",$0)
            gsub(/'\''/,"",$0)
            comment=$0
            cmd = "echo \""field"\" | perl -pe '\''s/(^|_)./uc($&)/ge;s/_//g'\'' "
            cmd | getline ufield
            close(cmd)
            idx += 1

            print "\t//" comment
            print "\t" ufield"\t"convertType(type)"\t`gorm:\"column:"field"\" json:\""field"\"`"
        }END{
            print "}\n"
            print "func (e *'${tableUpper}') TableName() string {"
            print "\treturn \"'${table}'\""
            print "}"
        }' > db_out/entity/${tableUpper}.go

    go fmt db_out/entity/${tableUpper}.go
    # 1.generate entity.proto
    ssed -n '/CREATE.*'${table}'`/,/ENGINE=/p' ${file} | \
    grep -v CREATE | grep -v ') ENGINE' | grep -v KEY | \
    awk -F " " '
            function convertType(ttype){
                if ( match(type,/bigint/) ){
                    return "int64"
                }
                if ( match(type,/tinyint|enum/) ){
                    return "int32"
                }
                if ( match(type,/int/) ){
                    return "int64"
                }
                 if ( match(type,/datetime|timestamp/) ){
                    return "int64"
                }
                if ( match(type,/varchar|text|char|blob|date/) ){
                    return "string"
                }

                if ( match(type,/decimal|double/) ){
                    return "double"
                }
                return ttype
            }
            BEGIN{
                idx=0
                print "syntax = \"proto3\";\n"
                print "package '${packagePrefix}'.'${table}';"
                print "option go_package =\"'${goPackagePrefix}'/'${table}'\";\n\n"
                print "//'${tableCmt}'"
                print "message '${tableUpper}' { "
            }{
            gsub(/`/,"",$1)
            field=$1
            type=$2
            gsub(/.*COMMENT/,"",$0)
            gsub(/'\''/,"",$0)
            comment=$0

            idx += 1

            print "\t//" comment
            print "\t"convertType(type)" "field" = "idx";"
        }END{
            print "}\n"
        }' > db_out/proto_entity/${table}.proto
done
