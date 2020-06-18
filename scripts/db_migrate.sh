#!/usr/bin/env bash
file=$1
packagePrefix=$2
goPackagePrefix=$3
#cat $file

rm -rf db_out
mkdir -p db_out/entity
mkdir -p db_out/proto_entity
go mod init .
table_list=($(grep CREATE ${file} | cut -d '`' -f 2))
function convert(){
    echo ${@}
}

tableUp_list=()
tableUp_pri_list=()
for table in ${table_list[@]}; do
    echo "proc ==== ${table}===="
    tableUpper=`echo "${table}" | perl -pe 's/(^|_)./uc($&)/ge;s/_//g'`
    tableUp_pri=`echo "${table}" | perl -pe 's/(_)./uc($&)/ge;s/_//g'`
    tableUp_list+=( ${tableUpper} )
    tableUp_pri_list+=( ${tableUp_pri} )

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
                withId=0
                print "package entity"

                print "import ("
                print "\t\"github.com/opay-org/lib-common/gorm_helper\""
                print ")"

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
            if ( field == "id" ) {
                withId = 1
            }
            print "\t//" comment
            print "\t" ufield"\t"convertType(type)"\t`gorm:\"column:"field"\" json:\""field"\"`"
        }END{
            print "}\n"
            print "func (e *'${tableUpper}') TableName() string {"
            print "\treturn \"'${table}'\""
            print "}\n"
            if ( withId > 0 ) {
                print "func (e *'${tableUpper}') PrimaryId() int64 {"
                print "\treturn e.Id "
                print "}"
                print "func (e *'${tableUpper}') Entity() gorm_helper.Entity {"
	            print "\treturn e"
                print "}"
            }

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

action_def_file=db_out/model_action_def.go

#######################
# generate model action def
cat > ${action_def_file} <<EOF
package main
import (
    "db_out/entity"

    "github.com/opay-org/lib-common/gorm_helper"
	"github.com/opay-org/lib-common/utils/obj_utils"
)
EOF
for table in ${table_list[@]}; do
    echo "proc ==== ${table}===="
    tableUpper=`echo "${table}" | perl -pe 's/(^|_)./uc($&)/ge;s/_//g'`

    cat >> ${action_def_file} <<EOF
var ${tableUpper}2MapWrapper = obj_utils.MustCompileCopyEntity2MapWrapper(entity.${tableUpper}{})
EOF
done
echo "" >> ${action_def_file}
for table in ${table_list[@]}; do
    echo "proc ==== ${table}===="
    tableUpper=`echo "${table}" | perl -pe 's/(^|_)./uc($&)/ge;s/_//g'`

    cat >>  ${action_def_file} <<EOF
type DataWriteAction${tableUpper} struct {
	*entity.${tableUpper}
	*gorm_helper.DataWriteActionBase
}
func (a *DataWriteAction${tableUpper}) Entity2MapWrapper() *obj_utils.CopyEntity2MapWrapper {
	return ${tableUpper}2MapWrapper
}
EOF
done

go fmt ${action_def_file}
#######################
# generate copy util
echo "#######################generate copy util#######################"
copy_util_file=db_out/copy_entity_utils.go
cat > ${copy_util_file} <<EOF
package main
import (
	"github.com/opay-org/lib-common/xlog"
	"github.com/opay-org/lib-common/utils/obj_utils"
)

var CopyEntityUtils = &copyEntityUtils{}

type copyEntityUtils struct {

EOF

for table in ${table_list[@]}; do
    tableUp_pri=`echo "${table}" | perl -pe 's/(_)./uc($&)/ge;s/_//g'`
    echo "${tableUp_pri}2e *obj_utils.CopyFieldWrapper" >> ${copy_util_file}
    echo "e2${tableUp_pri} *obj_utils.CopyFieldWrapper" >> ${copy_util_file}
done
echo "}" >> ${copy_util_file}
echo "func (l *copyEntityUtils) init() (err error) {" >> ${copy_util_file}
for table in ${table_list[@]}; do
    tableUpper=`echo "${table}" | perl -pe 's/(^|_)./uc($&)/ge;s/_//g'`
    tableUp_pri=`echo "${table}" | perl -pe 's/(_)./uc($&)/ge;s/_//g'`

    cat >> ${copy_util_file} <<EOF
    if l.${tableUp_pri}2e, err = obj_utils.CompileCopyFieldWrapper(${table}.${tableUpper}{},entity.${tableUpper}{}); err != nil {
        xlog.Error("l.${tableUp_pri}2e init failed||err=%v", err)
        return
    }
    if l.e2${tableUp_pri}, err = obj_utils.CompileCopyFieldWrapper(entity.${tableUpper}{},${table}.${tableUpper}{}); err != nil {
        xlog.Error("l.e2${tableUp_pri} init failed||err=%v", err)
        return
    }
EOF
done
echo "return }" >> ${copy_util_file}
for table in ${table_list[@]}; do
    tableUpper=`echo "${table}" | perl -pe 's/(^|_)./uc($&)/ge;s/_//g'`
    tableUp_pri=`echo "${table}" | perl -pe 's/(_)./uc($&)/ge;s/_//g'`

    cat >> ${copy_util_file} <<EOF

func (l *copyEntityUtils) ${tableUpper}2Entity(oSrc *${table}.${tableUpper}, eDest *entity.${tableUpper}, projection ...string) (skipFields []string, err error) {
	return l.${tableUp_pri}2e.CopyFieldValues(oSrc, eDest, projection...)
}

func (l *copyEntityUtils) Entity2${tableUpper}(eSrc *entity.${tableUpper}, oDest *${table}.${tableUpper}, projection ...string) (skipFields []string, err error) {
	return l.e2${tableUp_pri}.CopyFieldValues(eSrc, oDest, projection...)
}

EOF
done

go fmt ${copy_util_file}
copy_util_test_file=${copy_util_file}_test.go
echo "#######################generate copy util test#######################"

cat > ${copy_util_test_file} <<EOF
package main
import (
	"testing"
    "github.com/stretchr/testify/assert"
)

func Test_copyEntityUtilsInit(t *testing.T) {
    err := CopyEntityUtils.init()
    assert.Nil(t, err)
EOF

for tableUp_pri in ${tableUp_pri_list[@]}; do
    echo "	assert.Equal(t, len(CopyEntityUtils.e2${tableUp_pri}.SkippedFields()), 0, CopyEntityUtils.e2${tableUp_pri}.SkippedFields()) " >> ${copy_util_test_file}
    echo "	assert.Equal(t, len(CopyEntityUtils.${tableUp_pri}2e.SkippedFields()), 0, CopyEntityUtils.${tableUp_pri}2e.SkippedFields()) " >> ${copy_util_test_file}
done
echo "}" >> ${copy_util_test_file}

go fmt ${copy_util_test_file}