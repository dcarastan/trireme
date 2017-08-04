package remoteenforcer

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cnf/structhash"

	gomock "github.com/aporeto-inc/mock/gomock"
	"github.com/aporeto-inc/trireme/collector"
	"github.com/aporeto-inc/trireme/constants"
	"github.com/aporeto-inc/trireme/enforcer"
	"github.com/aporeto-inc/trireme/enforcer/utils/fqconfig"
	"github.com/aporeto-inc/trireme/enforcer/utils/rpcwrapper"
	"github.com/aporeto-inc/trireme/enforcer/utils/rpcwrapper/mock"
	"github.com/aporeto-inc/trireme/enforcer/utils/secrets"
	"github.com/aporeto-inc/trireme/policy"
	"github.com/aporeto-inc/trireme/supervisor"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	PrivatePEM []byte
	PublicPEM  []byte
	CAPem      []byte
	Token      []byte
)

func init() {
	PrivatePEM = []byte{0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x42, 0x45, 0x47, 0x49, 0x4E, 0x20, 0x45, 0x43, 0x20, 0x50, 0x52, 0x49, 0x56, 0x41, 0x54, 0x45, 0x20, 0x4B, 0x45, 0x59, 0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x0A, 0x4D, 0x48, 0x63, 0x43, 0x41, 0x51, 0x45, 0x45, 0x49, 0x4E, 0x41, 0x7A, 0x30, 0x70, 0x66, 0x46, 0x74, 0x31, 0x58, 0x35, 0x45, 0x71, 0x51, 0x61, 0x65, 0x6E, 0x6A, 0x6C, 0x38, 0x47, 0x70, 0x45, 0x75, 0x73, 0x72, 0x55, 0x2B, 0x70, 0x4E, 0x6C, 0x36, 0x59, 0x56, 0x48, 0x31, 0x35, 0x33, 0x4F, 0x33, 0x4E, 0x2B, 0x37, 0x6F, 0x41, 0x6F, 0x47, 0x43, 0x43, 0x71, 0x47, 0x53, 0x4D, 0x34, 0x39, 0x0A, 0x41, 0x77, 0x45, 0x48, 0x6F, 0x55, 0x51, 0x44, 0x51, 0x67, 0x41, 0x45, 0x4B, 0x4E, 0x4E, 0x57, 0x4F, 0x58, 0x31, 0x44, 0x53, 0x45, 0x70, 0x79, 0x62, 0x4A, 0x62, 0x64, 0x69, 0x4A, 0x68, 0x6A, 0x6B, 0x68, 0x72, 0x4C, 0x78, 0x75, 0x46, 0x72, 0x63, 0x4D, 0x49, 0x5A, 0x73, 0x61, 0x69, 0x58, 0x36, 0x77, 0x51, 0x66, 0x53, 0x54, 0x2B, 0x2B, 0x69, 0x46, 0x73, 0x74, 0x37, 0x68, 0x5A, 0x33, 0x0A, 0x36, 0x38, 0x6D, 0x43, 0x57, 0x75, 0x48, 0x41, 0x56, 0x6E, 0x51, 0x59, 0x35, 0x56, 0x64, 0x36, 0x74, 0x5A, 0x34, 0x51, 0x6E, 0x54, 0x54, 0x65, 0x79, 0x4D, 0x45, 0x61, 0x79, 0x52, 0x65, 0x70, 0x4B, 0x51, 0x3D, 0x3D, 0x0A, 0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x45, 0x4E, 0x44, 0x20, 0x45, 0x43, 0x20, 0x50, 0x52, 0x49, 0x56, 0x41, 0x54, 0x45, 0x20, 0x4B, 0x45, 0x59, 0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x0A}

	PublicPEM = []byte{0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x42, 0x45, 0x47, 0x49, 0x4E, 0x20, 0x43, 0x45, 0x52, 0x54, 0x49, 0x46, 0x49, 0x43, 0x41, 0x54, 0x45, 0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x0A, 0x4D, 0x49, 0x49, 0x43, 0x50, 0x7A, 0x43, 0x43, 0x41, 0x65, 0x61, 0x67, 0x41, 0x77, 0x49, 0x42, 0x41, 0x67, 0x49, 0x51, 0x64, 0x48, 0x32, 0x53, 0x5A, 0x39, 0x50, 0x6B, 0x58, 0x63, 0x46, 0x59, 0x68, 0x4C, 0x7A, 0x35, 0x2F, 0x6C, 0x58, 0x4F, 0x6F, 0x44, 0x41, 0x4B, 0x42, 0x67, 0x67, 0x71, 0x68, 0x6B, 0x6A, 0x4F, 0x50, 0x51, 0x51, 0x44, 0x41, 0x6A, 0x42, 0x5A, 0x4D, 0x51, 0x73, 0x77, 0x0A, 0x43, 0x51, 0x59, 0x44, 0x56, 0x51, 0x51, 0x47, 0x45, 0x77, 0x4A, 0x56, 0x55, 0x7A, 0x45, 0x4C, 0x4D, 0x41, 0x6B, 0x47, 0x41, 0x31, 0x55, 0x45, 0x43, 0x41, 0x77, 0x43, 0x51, 0x30, 0x45, 0x78, 0x45, 0x54, 0x41, 0x50, 0x42, 0x67, 0x4E, 0x56, 0x42, 0x41, 0x63, 0x4D, 0x43, 0x46, 0x4E, 0x68, 0x62, 0x69, 0x42, 0x4B, 0x62, 0x33, 0x4E, 0x6C, 0x4D, 0x52, 0x41, 0x77, 0x44, 0x67, 0x59, 0x44, 0x0A, 0x56, 0x51, 0x51, 0x4B, 0x44, 0x41, 0x64, 0x42, 0x63, 0x47, 0x39, 0x79, 0x5A, 0x58, 0x52, 0x76, 0x4D, 0x52, 0x67, 0x77, 0x46, 0x67, 0x59, 0x44, 0x56, 0x51, 0x51, 0x44, 0x44, 0x41, 0x39, 0x42, 0x63, 0x47, 0x39, 0x79, 0x5A, 0x58, 0x52, 0x76, 0x49, 0x46, 0x4A, 0x76, 0x62, 0x33, 0x51, 0x67, 0x51, 0x30, 0x45, 0x77, 0x48, 0x68, 0x63, 0x4E, 0x4D, 0x54, 0x63, 0x77, 0x4F, 0x44, 0x41, 0x79, 0x0A, 0x4D, 0x6A, 0x41, 0x7A, 0x4D, 0x54, 0x55, 0x79, 0x57, 0x68, 0x63, 0x4E, 0x4D, 0x54, 0x67, 0x77, 0x4F, 0x44, 0x41, 0x79, 0x4D, 0x6A, 0x41, 0x7A, 0x4D, 0x54, 0x55, 0x79, 0x57, 0x6A, 0x42, 0x67, 0x4D, 0x52, 0x4D, 0x77, 0x45, 0x51, 0x59, 0x44, 0x56, 0x51, 0x51, 0x4B, 0x45, 0x77, 0x70, 0x7A, 0x61, 0x57, 0x4A, 0x70, 0x59, 0x32, 0x56, 0x75, 0x64, 0x47, 0x39, 0x7A, 0x4D, 0x52, 0x6F, 0x77, 0x0A, 0x47, 0x41, 0x59, 0x44, 0x56, 0x51, 0x51, 0x4C, 0x45, 0x78, 0x46, 0x68, 0x63, 0x47, 0x39, 0x79, 0x5A, 0x58, 0x52, 0x76, 0x4C, 0x57, 0x56, 0x75, 0x5A, 0x6D, 0x39, 0x79, 0x59, 0x32, 0x56, 0x79, 0x5A, 0x44, 0x45, 0x74, 0x4D, 0x43, 0x73, 0x47, 0x41, 0x31, 0x55, 0x45, 0x41, 0x77, 0x77, 0x6B, 0x4E, 0x54, 0x6B, 0x34, 0x4D, 0x6A, 0x4D, 0x32, 0x59, 0x6A, 0x67, 0x78, 0x59, 0x7A, 0x49, 0x31, 0x0A, 0x4D, 0x6D, 0x4D, 0x77, 0x4D, 0x44, 0x41, 0x78, 0x4D, 0x44, 0x49, 0x32, 0x4E, 0x6A, 0x56, 0x6B, 0x51, 0x43, 0x39, 0x7A, 0x61, 0x57, 0x4A, 0x70, 0x59, 0x32, 0x56, 0x75, 0x64, 0x47, 0x39, 0x7A, 0x4D, 0x46, 0x6B, 0x77, 0x45, 0x77, 0x59, 0x48, 0x4B, 0x6F, 0x5A, 0x49, 0x7A, 0x6A, 0x30, 0x43, 0x41, 0x51, 0x59, 0x49, 0x4B, 0x6F, 0x5A, 0x49, 0x7A, 0x6A, 0x30, 0x44, 0x41, 0x51, 0x63, 0x44, 0x0A, 0x51, 0x67, 0x41, 0x45, 0x4B, 0x4E, 0x4E, 0x57, 0x4F, 0x58, 0x31, 0x44, 0x53, 0x45, 0x70, 0x79, 0x62, 0x4A, 0x62, 0x64, 0x69, 0x4A, 0x68, 0x6A, 0x6B, 0x68, 0x72, 0x4C, 0x78, 0x75, 0x46, 0x72, 0x63, 0x4D, 0x49, 0x5A, 0x73, 0x61, 0x69, 0x58, 0x36, 0x77, 0x51, 0x66, 0x53, 0x54, 0x2B, 0x2B, 0x69, 0x46, 0x73, 0x74, 0x37, 0x68, 0x5A, 0x33, 0x36, 0x38, 0x6D, 0x43, 0x57, 0x75, 0x48, 0x41, 0x0A, 0x56, 0x6E, 0x51, 0x59, 0x35, 0x56, 0x64, 0x36, 0x74, 0x5A, 0x34, 0x51, 0x6E, 0x54, 0x54, 0x65, 0x79, 0x4D, 0x45, 0x61, 0x79, 0x52, 0x65, 0x70, 0x4B, 0x61, 0x4F, 0x42, 0x69, 0x44, 0x43, 0x42, 0x68, 0x54, 0x41, 0x64, 0x42, 0x67, 0x4E, 0x56, 0x48, 0x53, 0x55, 0x45, 0x46, 0x6A, 0x41, 0x55, 0x42, 0x67, 0x67, 0x72, 0x42, 0x67, 0x45, 0x46, 0x42, 0x51, 0x63, 0x44, 0x41, 0x51, 0x59, 0x49, 0x0A, 0x4B, 0x77, 0x59, 0x42, 0x42, 0x51, 0x55, 0x48, 0x41, 0x77, 0x49, 0x77, 0x44, 0x41, 0x59, 0x44, 0x56, 0x52, 0x30, 0x54, 0x41, 0x51, 0x48, 0x2F, 0x42, 0x41, 0x49, 0x77, 0x41, 0x44, 0x41, 0x66, 0x42, 0x67, 0x4E, 0x56, 0x48, 0x53, 0x4D, 0x45, 0x47, 0x44, 0x41, 0x57, 0x67, 0x42, 0x51, 0x6D, 0x77, 0x6E, 0x73, 0x54, 0x45, 0x49, 0x68, 0x7A, 0x41, 0x70, 0x4D, 0x50, 0x4C, 0x43, 0x58, 0x37, 0x0A, 0x4C, 0x30, 0x6E, 0x4C, 0x63, 0x70, 0x6D, 0x78, 0x7A, 0x44, 0x41, 0x31, 0x42, 0x67, 0x4E, 0x56, 0x48, 0x52, 0x45, 0x45, 0x4C, 0x6A, 0x41, 0x73, 0x67, 0x67, 0x39, 0x7A, 0x61, 0x57, 0x4A, 0x70, 0x4C, 0x56, 0x5A, 0x70, 0x63, 0x6E, 0x52, 0x31, 0x59, 0x57, 0x78, 0x43, 0x62, 0x33, 0x69, 0x42, 0x47, 0x57, 0x56, 0x75, 0x5A, 0x6D, 0x39, 0x79, 0x59, 0x32, 0x56, 0x79, 0x5A, 0x45, 0x42, 0x7A, 0x0A, 0x61, 0x57, 0x4A, 0x70, 0x4C, 0x56, 0x5A, 0x70, 0x63, 0x6E, 0x52, 0x31, 0x59, 0x57, 0x78, 0x43, 0x62, 0x33, 0x67, 0x77, 0x43, 0x67, 0x59, 0x49, 0x4B, 0x6F, 0x5A, 0x49, 0x7A, 0x6A, 0x30, 0x45, 0x41, 0x77, 0x49, 0x44, 0x52, 0x77, 0x41, 0x77, 0x52, 0x41, 0x49, 0x67, 0x4E, 0x45, 0x4C, 0x4A, 0x61, 0x31, 0x37, 0x56, 0x74, 0x6B, 0x61, 0x75, 0x76, 0x49, 0x54, 0x75, 0x67, 0x2F, 0x6C, 0x74, 0x0A, 0x71, 0x39, 0x53, 0x4D, 0x45, 0x50, 0x53, 0x52, 0x52, 0x66, 0x33, 0x79, 0x46, 0x4C, 0x49, 0x47, 0x75, 0x77, 0x5A, 0x4A, 0x2B, 0x73, 0x49, 0x43, 0x49, 0x46, 0x4A, 0x6C, 0x42, 0x36, 0x50, 0x46, 0x6A, 0x43, 0x62, 0x44, 0x79, 0x73, 0x6F, 0x38, 0x56, 0x7A, 0x45, 0x4D, 0x78, 0x75, 0x56, 0x76, 0x4A, 0x77, 0x66, 0x6B, 0x4C, 0x6D, 0x7A, 0x68, 0x66, 0x63, 0x35, 0x6E, 0x58, 0x4E, 0x47, 0x6F, 0x0A, 0x66, 0x48, 0x61, 0x6E, 0x0A, 0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x45, 0x4E, 0x44, 0x20, 0x43, 0x45, 0x52, 0x54, 0x49, 0x46, 0x49, 0x43, 0x41, 0x54, 0x45, 0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x0A}

	CAPem = []byte{0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x42, 0x45, 0x47, 0x49, 0x4E, 0x20, 0x43, 0x45, 0x52, 0x54, 0x49, 0x46, 0x49, 0x43, 0x41, 0x54, 0x45, 0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x0A, 0x4D, 0x49, 0x49, 0x42, 0x2B, 0x6A, 0x43, 0x43, 0x41, 0x5A, 0x2B, 0x67, 0x41, 0x77, 0x49, 0x42, 0x41, 0x67, 0x49, 0x4A, 0x41, 0x4E, 0x51, 0x77, 0x4D, 0x73, 0x67, 0x53, 0x67, 0x46, 0x36, 0x79, 0x4D, 0x41, 0x6F, 0x47, 0x43, 0x43, 0x71, 0x47, 0x53, 0x4D, 0x34, 0x39, 0x42, 0x41, 0x4D, 0x44, 0x4D, 0x46, 0x6B, 0x78, 0x43, 0x7A, 0x41, 0x4A, 0x42, 0x67, 0x4E, 0x56, 0x42, 0x41, 0x59, 0x54, 0x0A, 0x41, 0x6C, 0x56, 0x54, 0x4D, 0x51, 0x73, 0x77, 0x43, 0x51, 0x59, 0x44, 0x56, 0x51, 0x51, 0x49, 0x44, 0x41, 0x4A, 0x44, 0x51, 0x54, 0x45, 0x52, 0x4D, 0x41, 0x38, 0x47, 0x41, 0x31, 0x55, 0x45, 0x42, 0x77, 0x77, 0x49, 0x55, 0x32, 0x46, 0x75, 0x49, 0x45, 0x70, 0x76, 0x63, 0x32, 0x55, 0x78, 0x45, 0x44, 0x41, 0x4F, 0x42, 0x67, 0x4E, 0x56, 0x42, 0x41, 0x6F, 0x4D, 0x42, 0x30, 0x46, 0x77, 0x0A, 0x62, 0x33, 0x4A, 0x6C, 0x64, 0x47, 0x38, 0x78, 0x47, 0x44, 0x41, 0x57, 0x42, 0x67, 0x4E, 0x56, 0x42, 0x41, 0x4D, 0x4D, 0x44, 0x30, 0x46, 0x77, 0x62, 0x33, 0x4A, 0x6C, 0x64, 0x47, 0x38, 0x67, 0x55, 0x6D, 0x39, 0x76, 0x64, 0x43, 0x42, 0x44, 0x51, 0x54, 0x41, 0x65, 0x46, 0x77, 0x30, 0x78, 0x4E, 0x7A, 0x41, 0x7A, 0x4D, 0x6A, 0x49, 0x79, 0x4D, 0x44, 0x49, 0x35, 0x4D, 0x6A, 0x46, 0x61, 0x0A, 0x46, 0x77, 0x30, 0x79, 0x4E, 0x7A, 0x41, 0x7A, 0x4D, 0x6A, 0x41, 0x79, 0x4D, 0x44, 0x49, 0x35, 0x4D, 0x6A, 0x46, 0x61, 0x4D, 0x46, 0x6B, 0x78, 0x43, 0x7A, 0x41, 0x4A, 0x42, 0x67, 0x4E, 0x56, 0x42, 0x41, 0x59, 0x54, 0x41, 0x6C, 0x56, 0x54, 0x4D, 0x51, 0x73, 0x77, 0x43, 0x51, 0x59, 0x44, 0x56, 0x51, 0x51, 0x49, 0x44, 0x41, 0x4A, 0x44, 0x51, 0x54, 0x45, 0x52, 0x4D, 0x41, 0x38, 0x47, 0x0A, 0x41, 0x31, 0x55, 0x45, 0x42, 0x77, 0x77, 0x49, 0x55, 0x32, 0x46, 0x75, 0x49, 0x45, 0x70, 0x76, 0x63, 0x32, 0x55, 0x78, 0x45, 0x44, 0x41, 0x4F, 0x42, 0x67, 0x4E, 0x56, 0x42, 0x41, 0x6F, 0x4D, 0x42, 0x30, 0x46, 0x77, 0x62, 0x33, 0x4A, 0x6C, 0x64, 0x47, 0x38, 0x78, 0x47, 0x44, 0x41, 0x57, 0x42, 0x67, 0x4E, 0x56, 0x42, 0x41, 0x4D, 0x4D, 0x44, 0x30, 0x46, 0x77, 0x62, 0x33, 0x4A, 0x6C, 0x0A, 0x64, 0x47, 0x38, 0x67, 0x55, 0x6D, 0x39, 0x76, 0x64, 0x43, 0x42, 0x44, 0x51, 0x54, 0x42, 0x5A, 0x4D, 0x42, 0x4D, 0x47, 0x42, 0x79, 0x71, 0x47, 0x53, 0x4D, 0x34, 0x39, 0x41, 0x67, 0x45, 0x47, 0x43, 0x43, 0x71, 0x47, 0x53, 0x4D, 0x34, 0x39, 0x41, 0x77, 0x45, 0x48, 0x41, 0x30, 0x49, 0x41, 0x42, 0x4D, 0x32, 0x37, 0x39, 0x4B, 0x48, 0x66, 0x6E, 0x37, 0x6E, 0x6E, 0x61, 0x75, 0x30, 0x6D, 0x0A, 0x79, 0x58, 0x59, 0x34, 0x33, 0x4D, 0x32, 0x4E, 0x74, 0x4C, 0x63, 0x6E, 0x78, 0x73, 0x53, 0x58, 0x52, 0x5A, 0x47, 0x56, 0x6E, 0x6B, 0x31, 0x45, 0x74, 0x53, 0x32, 0x4A, 0x49, 0x73, 0x46, 0x39, 0x38, 0x2B, 0x52, 0x38, 0x38, 0x52, 0x2B, 0x30, 0x7A, 0x5A, 0x55, 0x73, 0x70, 0x52, 0x30, 0x57, 0x7A, 0x71, 0x6E, 0x4C, 0x76, 0x69, 0x43, 0x49, 0x44, 0x37, 0x4E, 0x65, 0x79, 0x79, 0x30, 0x6D, 0x0A, 0x53, 0x6A, 0x59, 0x63, 0x71, 0x54, 0x43, 0x6A, 0x55, 0x44, 0x42, 0x4F, 0x4D, 0x42, 0x30, 0x47, 0x41, 0x31, 0x55, 0x64, 0x44, 0x67, 0x51, 0x57, 0x42, 0x42, 0x51, 0x6D, 0x77, 0x6E, 0x73, 0x54, 0x45, 0x49, 0x68, 0x7A, 0x41, 0x70, 0x4D, 0x50, 0x4C, 0x43, 0x58, 0x37, 0x4C, 0x30, 0x6E, 0x4C, 0x63, 0x70, 0x6D, 0x78, 0x7A, 0x44, 0x41, 0x66, 0x42, 0x67, 0x4E, 0x56, 0x48, 0x53, 0x4D, 0x45, 0x0A, 0x47, 0x44, 0x41, 0x57, 0x67, 0x42, 0x51, 0x6D, 0x77, 0x6E, 0x73, 0x54, 0x45, 0x49, 0x68, 0x7A, 0x41, 0x70, 0x4D, 0x50, 0x4C, 0x43, 0x58, 0x37, 0x4C, 0x30, 0x6E, 0x4C, 0x63, 0x70, 0x6D, 0x78, 0x7A, 0x44, 0x41, 0x4D, 0x42, 0x67, 0x4E, 0x56, 0x48, 0x52, 0x4D, 0x45, 0x42, 0x54, 0x41, 0x44, 0x41, 0x51, 0x48, 0x2F, 0x4D, 0x41, 0x6F, 0x47, 0x43, 0x43, 0x71, 0x47, 0x53, 0x4D, 0x34, 0x39, 0x0A, 0x42, 0x41, 0x4D, 0x44, 0x41, 0x30, 0x6B, 0x41, 0x4D, 0x45, 0x59, 0x43, 0x49, 0x51, 0x43, 0x31, 0x65, 0x53, 0x45, 0x65, 0x33, 0x32, 0x6A, 0x54, 0x37, 0x62, 0x75, 0x79, 0x2B, 0x58, 0x4B, 0x51, 0x4E, 0x31, 0x71, 0x70, 0x39, 0x62, 0x69, 0x70, 0x4C, 0x43, 0x71, 0x36, 0x4F, 0x66, 0x64, 0x4C, 0x2B, 0x52, 0x70, 0x51, 0x38, 0x4F, 0x63, 0x2B, 0x6D, 0x77, 0x49, 0x68, 0x41, 0x4F, 0x59, 0x4F, 0x0A, 0x67, 0x4E, 0x62, 0x7A, 0x6D, 0x46, 0x35, 0x2F, 0x51, 0x53, 0x44, 0x6A, 0x79, 0x33, 0x47, 0x69, 0x50, 0x4F, 0x2F, 0x39, 0x6B, 0x74, 0x59, 0x78, 0x41, 0x44, 0x30, 0x65, 0x49, 0x32, 0x49, 0x6A, 0x73, 0x50, 0x6C, 0x42, 0x6D, 0x6C, 0x73, 0x67, 0x0A, 0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x45, 0x4E, 0x44, 0x20, 0x43, 0x45, 0x52, 0x54, 0x49, 0x46, 0x49, 0x43, 0x41, 0x54, 0x45, 0x2D, 0x2D, 0x2D, 0x2D, 0x2D, 0x0A}

	Token = []byte{0x65, 0x79, 0x4A, 0x68, 0x62, 0x47, 0x63, 0x69, 0x4F, 0x69, 0x4A, 0x46, 0x55, 0x7A, 0x49, 0x31, 0x4E, 0x69, 0x49, 0x73, 0x49, 0x6E, 0x52, 0x35, 0x63, 0x43, 0x49, 0x36, 0x49, 0x6B, 0x70, 0x58, 0x56, 0x43, 0x4A, 0x39, 0x2E, 0x65, 0x79, 0x4A, 0x59, 0x49, 0x6A, 0x6F, 0x78, 0x4F, 0x44, 0x51, 0x32, 0x4E, 0x54, 0x6B, 0x78, 0x4D, 0x7A, 0x63, 0x33, 0x4E, 0x44, 0x41, 0x35, 0x4D, 0x7A, 0x4D, 0x35, 0x4D, 0x7A, 0x51, 0x33, 0x4D, 0x54, 0x4D, 0x77, 0x4D, 0x6A, 0x4D, 0x35, 0x4E, 0x6A, 0x45, 0x79, 0x4E, 0x6A, 0x55, 0x79, 0x4D, 0x7A, 0x45, 0x77, 0x4E, 0x44, 0x51, 0x30, 0x4F, 0x44, 0x63, 0x34, 0x4D, 0x7A, 0x45, 0x78, 0x4E, 0x6A, 0x4D, 0x30, 0x4E, 0x7A, 0x6B, 0x32, 0x4D, 0x6A, 0x4D, 0x32, 0x4D, 0x7A, 0x67, 0x30, 0x4E, 0x54, 0x59, 0x30, 0x4E, 0x6A, 0x51, 0x78, 0x4E, 0x7A, 0x67, 0x78, 0x4E, 0x44, 0x41, 0x78, 0x4F, 0x44, 0x63, 0x35, 0x4F, 0x44, 0x4D, 0x30, 0x4D, 0x44, 0x51, 0x78, 0x4E, 0x53, 0x77, 0x69, 0x57, 0x53, 0x49, 0x36, 0x4F, 0x44, 0x59, 0x78, 0x4F, 0x44, 0x41, 0x7A, 0x4E, 0x6A, 0x45, 0x33, 0x4D, 0x6A, 0x67, 0x34, 0x4D, 0x54, 0x6B, 0x79, 0x4D, 0x44, 0x41, 0x30, 0x4D, 0x6A, 0x41, 0x33, 0x4D, 0x44, 0x63, 0x30, 0x4D, 0x44, 0x6B, 0x78, 0x4D, 0x54, 0x41, 0x33, 0x4D, 0x54, 0x49, 0x33, 0x4D, 0x7A, 0x49, 0x78, 0x4F, 0x54, 0x45, 0x34, 0x4D, 0x54, 0x45, 0x77, 0x4F, 0x44, 0x41, 0x77, 0x4E, 0x54, 0x41, 0x79, 0x4F, 0x54, 0x59, 0x79, 0x4D, 0x6A, 0x49, 0x78, 0x4D, 0x54, 0x41, 0x32, 0x4E, 0x44, 0x41, 0x30, 0x4D, 0x54, 0x6B, 0x32, 0x4F, 0x54, 0x49, 0x34, 0x4D, 0x54, 0x55, 0x78, 0x4D, 0x6A, 0x55, 0x31, 0x4E, 0x54, 0x55, 0x30, 0x4F, 0x54, 0x63, 0x73, 0x49, 0x6D, 0x56, 0x34, 0x63, 0x43, 0x49, 0x36, 0x4D, 0x54, 0x55, 0x7A, 0x4D, 0x7A, 0x49, 0x30, 0x4D, 0x54, 0x6B, 0x78, 0x4D, 0x6E, 0x30, 0x2E, 0x56, 0x43, 0x44, 0x30, 0x54, 0x61, 0x4C, 0x69, 0x66, 0x74, 0x35, 0x63, 0x6A, 0x6E, 0x66, 0x74, 0x73, 0x7A, 0x57, 0x63, 0x43, 0x74, 0x56, 0x64, 0x59, 0x49, 0x63, 0x5A, 0x44, 0x58, 0x63, 0x73, 0x67, 0x66, 0x47, 0x41, 0x69, 0x33, 0x42, 0x77, 0x6F, 0x73, 0x4A, 0x50, 0x68, 0x6F, 0x76, 0x6A, 0x57, 0x65, 0x56, 0x65, 0x74, 0x6E, 0x55, 0x44, 0x44, 0x46, 0x69, 0x45, 0x37, 0x4E, 0x78, 0x76, 0x4E, 0x6A, 0x32, 0x52, 0x43, 0x53, 0x79, 0x4A, 0x76, 0x2D, 0x52, 0x6F, 0x71, 0x72, 0x6F, 0x78, 0x4E, 0x48, 0x4B, 0x4B, 0x37, 0x77}
}

func initTestEnfReqPayload() rpcwrapper.InitRequestPayload {
	var initEnfPayload rpcwrapper.InitRequestPayload

	dur, _ := time.ParseDuration("8760h0m0s")
	initEnfPayload.Validity = dur
	initEnfPayload.MutualAuth = true
	initEnfPayload.SecretType = 2
	initEnfPayload.ServerID = "598236b81c252c000102665d"
	initEnfPayload.FqConfig = filterQ()
	initEnfPayload.PrivatePEM = PrivatePEM
	initEnfPayload.PublicPEM = PublicPEM
	initEnfPayload.CAPEM = CAPem
	initEnfPayload.Token = Token

	return initEnfPayload
}

func filterQ() *fqconfig.FilterQueue {
	var initFilterQ fqconfig.FilterQueue

	initFilterQ.QueueSeparation = false
	initFilterQ.MarkValue = 1000
	initFilterQ.NetworkQueue = 4
	initFilterQ.NumberOfApplicationQueues = 4
	initFilterQ.NumberOfNetworkQueues = 4
	initFilterQ.ApplicationQueue = 0
	initFilterQ.ApplicationQueueSize = 500
	initFilterQ.NetworkQueueSize = 500
	initFilterQ.NetworkQueuesSynStr = "4:7"
	initFilterQ.NetworkQueuesAckStr = "4:7"
	initFilterQ.NetworkQueuesSynAckStr = "4:7"
	initFilterQ.NetworkQueuesSvcStr = "4:7"
	initFilterQ.ApplicationQueuesSynStr = "0:3"
	initFilterQ.ApplicationQueuesAckStr = "0:3"
	initFilterQ.ApplicationQueuesSvcStr = "0:3"
	initFilterQ.ApplicationQueuesSynAckStr = "0:3"

	return &initFilterQ
}

func initTestSupReqPayload() rpcwrapper.InitSupervisorPayload {
	var initSupPayload rpcwrapper.InitSupervisorPayload

	initSupPayload.TriremeNetworks = []string{"127.0.0.1/32 172.0.0.0/8 10.0.0.0/8"}

	return initSupPayload
}

func initIdentity(id string) *policy.TagStore {
	var initID policy.TagStore

	initID.Tags = []string{id}

	return &initID
}

func initAnnotations(an string) *policy.TagStore {
	var initAnno policy.TagStore

	initAnno.Tags = []string{an}

	return &initAnno
}

func initTrans() policy.TagSelectorList {
	var tags policy.TagSelectorList
	var tag policy.TagSelector
	var keyval policy.KeyValueOperator

	keyval.Key = "@usr:role"
	keyval.Value = []string{"server"}
	keyval.Operator = "="
	tag.Clause = []policy.KeyValueOperator{keyval}
	tag.Action = 1
	tags = []policy.TagSelector{tag}

	return tags
}

func initTestSupPayload() rpcwrapper.SuperviseRequestPayload {

	var initPayload rpcwrapper.SuperviseRequestPayload
  idString:="$namespace=/sibicentos @usr:role=client AporetoContextID=59812ccc27b430000135fbf3"
  anoString:="@sys:name=/nervous_hermann @usr:role=client @usr:vendor=CentOS $id=59812ccc27b430000135fbf3 $namespace=/sibicentos @usr:build-date=20170705 @usr:license=GPLv2 @usr:name=CentOS Base Image $nativecontextid=ac0d3577e808 $operationalstatus=Running role=client $id=59812ccc27b430000135fbf3 $identity=processingunit $namespace=/sibicentos $protected=false $type=Docker @sys:image=centos @usr:role=client $description=centos $enforcerid=598236b81c252c000102665d $name=centos $id=59812ccc27b430000135fbf3 $namespace=/sibicentos"

	initPayload.ContextID = "ac0d3577e808"
	initPayload.ManagementID = "59812ccc27b430000135fbf3"
	initPayload.TriremeAction = 2
	//initPayload.ApplicationACLs
	//initPayload.NetworkACLs
	initPayload.PolicyIPs = policy.ExtendedMap{"bridge": "172.17.0.2"}
	initPayload.Identity = initIdentity(idString)
	initPayload.Annotations = initAnnotations(anoString)
	//initPayload.ReceiverRules
	initPayload.TransmitterRules = initTrans()
	//initPayload.ExcludedNetworks=
	initPayload.TriremeNetworks = []string{"127.0.0.1/32 172.0.0.0/8 10.0.0.0/8"}

	return initPayload
}

func initTestEnfPayload() rpcwrapper.EnforcePayload{

  var initPayload rpcwrapper.EnforcePayload
  idString:="@usr:role=client $namespace=/sibicentos AporetoContextID=5983bc8c923caa0001337b11"
  anoString:="@sys:name=/inspiring_roentgen $namespace=/sibicentos @usr:build-date=20170801 @usr:license=GPLv2 @usr:name=CentOS Base Image @usr:role=client @usr:vendor=CentOS $id=5983bc8c923caa0001337b11 $namespace=/sibicentos $operationalstatus=Running $protected=false $type=Docker $description=centos $enforcerid=5983bba4923caa0001337a19 $name=centos $nativecontextid=b06f47830f64 @sys:image=centos @usr:role=client role=client $id=5983bc8c923caa0001337b11 $identity=processingunit $id=5983bc8c923caa0001337b11 $namespace=/sibicentos"

  initPayload.ContextID = "b06f47830f64"
  initPayload.ManagementID = "5983bc8c923caa0001337b11"
  initPayload.TriremeAction = 2
  //initPayload.ApplicationACLs
  //initPayload.NetworkACLs
  initPayload.PolicyIPs = policy.ExtendedMap{"bridge": "172.17.0.2"}
  initPayload.Identity = initIdentity(idString)
  initPayload.Annotations = initAnnotations(anoString)
  //initPayload.ReceiverRules
  initPayload.TransmitterRules = initTrans()
  //initPayload.ExcludedNetworks=
  initPayload.TriremeNetworks = []string{"127.0.0.1/32 172.0.0.0/8 10.0.0.0/8"}

  return initPayload
}

func initTestUnEnfPayload() rpcwrapper.UnEnforcePayload{

  var initPayload rpcwrapper.UnEnforcePayload

  initPayload.ContextID = "b06f47830f64"

  return initPayload
}

func initTestUnSupPayload() rpcwrapper.UnSupervisePayload{

  var initPayload rpcwrapper.UnSupervisePayload

  initPayload.ContextID = "ac0d3577e808"

  return initPayload
}

func TestNewServer(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("When I try to retrieve rpc server handle", t, func() {
		rpcHdl := mock_rpcwrapper.NewMockRPCServer(ctrl)

		Convey("Then rpcHdl should resemble rpcwrapper struct", func() {
			So(rpcHdl, ShouldNotBeNil)
		})

		Convey("When I try to create new server with no env set", func() {
			rpcHdl.EXPECT().StartServer(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			var service enforcer.PacketProcessor
			pcchan := "/tmp/test.sock"
			secret := "mysecret"
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get error for no stats", func() {
				So(server, ShouldBeNil)
				So(err, ShouldResemble, fmt.Errorf("No path to stats socket provided"))
			})
		})

		Convey("When I try to create new server with env set", func() {
			os.Setenv("STATSCHANNEL_PATH", "/tmp/test.sock")
			os.Setenv("STATS_SECRET", "mysecret")
			var service enforcer.PacketProcessor
			pcchan := os.Getenv("STATSCHANNEL_PATH")
			secret := os.Getenv("STATS_SECRET")
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get no error", func() {
				So(server, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
			os.Setenv("STATSCHANNEL_PATH", "")
			os.Setenv("STATS_SECRET", "")
		})
	})
}

func TestInitEnforcer(t *testing.T) {

	if os.Getenv("USER") != "root" {
		t.SkipNow()
	}

	Convey("When I try to retrieve rpc server handle", t, func() {
		rpcHdl := rpcwrapper.NewRPCServer()
		var rpcWrpper rpcwrapper.RPCWrapper

		Convey("Then rpcHdl should resemble rpcwrapper struct", func() {
			So(rpcHdl, ShouldResemble, &rpcWrpper)
		})

		Convey("When I try to create new server with env set", func() {
			os.Setenv("STATSCHANNEL_PATH", "/tmp/test.sock")
			os.Setenv("STATS_SECRET", "T6UYZGcKW-aum_vi-XakafF3vHV7F6x8wdofZs7akGU=")
			var service enforcer.PacketProcessor
			pcchan := os.Getenv("STATSCHANNEL_PATH")
			secret := os.Getenv("STATS_SECRET")
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get no error", func() {
				So(server, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("When I try to initiate an enforcer", func() {
				var rpcwrperreq rpcwrapper.Request
				var rpcwrperres rpcwrapper.Response

				digest := hmac.New(sha256.New, []byte(os.Getenv("STATS_SECRET")))
				if _, err := digest.Write(structhash.Dump(rpcwrperreq.Payload, 1)); err != nil {
					So(err, ShouldBeNil)
				}
				rpcwrperreq.HashAuth = digest.Sum(nil)

				rpcwrperreq.HashAuth = []byte{0xC5, 0xD1, 0x24, 0x36, 0x1A, 0xFC, 0x66, 0x3E, 0xAE, 0xD7, 0x68, 0xCE, 0x88, 0x72, 0xC0, 0x97, 0xE4, 0x27, 0x70, 0x6C, 0x47, 0x31, 0x67, 0xEF, 0xD5, 0xCE, 0x73, 0x99, 0x7B, 0xAC, 0x25, 0x94}
				rpcwrperreq.Payload = initTestEnfReqPayload()
        rpcwrperres.Status = ""

				//err := server.InitEnforcer(rpcwrperreq, &rpcwrperres)

				Convey("Then I should get no error", func() {
				//	So(err, ShouldBeNil)
				})
				os.Setenv("STATSCHANNEL_PATH", "")
				os.Setenv("STATS_SECRET", "")
			})
		})
	})
}

func TestInitSupervisor(t *testing.T) {

	Convey("When I try to retrieve rpc server handle", t, func() {

		rpcHdl := rpcwrapper.NewRPCServer()
		var rpcWrpper rpcwrapper.RPCWrapper

		Convey("Then rpcHdl should resemble rpcwrapper struct", func() {
			So(rpcHdl, ShouldResemble, &rpcWrpper)
		})

		Convey("When I try to create new server with env set", func() {
			os.Setenv("STATSCHANNEL_PATH", "/tmp/test.sock")
			os.Setenv("STATS_SECRET", "n1KroWMWKP8nJnpWfwSsQu855yvP-ZPaNr-TJFl3gzM=")
			var service enforcer.PacketProcessor
			pcchan := os.Getenv("STATSCHANNEL_PATH")
			secret := os.Getenv("STATS_SECRET")
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get no error", func() {
				So(server, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("When I try to initiate the supervisor with no enforcer", func() {
				var rpcwrperreq rpcwrapper.Request
				var rpcwrperres rpcwrapper.Response

				rpcwrperreq.HashAuth = []byte{0x47, 0xBE, 0x1A, 0x01, 0x47, 0x4F, 0x4A, 0x7A, 0xB5, 0xDA, 0x97, 0x46, 0xF3, 0x98, 0x50, 0x86, 0xB1, 0xF7, 0x05, 0x65, 0x6F, 0x58, 0x8C, 0x2C, 0x23, 0x9B, 0xA2, 0x82, 0x40, 0x45, 0x24, 0x45}
				rpcwrperreq.Payload = initTestSupReqPayload()
				rpcwrperres.Status = ""

				digest := hmac.New(sha256.New, []byte(os.Getenv("STATS_SECRET")))
				if _, err := digest.Write(structhash.Dump(rpcwrperreq.Payload, 1)); err != nil {
					So(err, ShouldBeNil)
				}
				rpcwrperreq.HashAuth = digest.Sum(nil)

				err := server.InitSupervisor(rpcwrperreq, &rpcwrperres)

				Convey("Then I should get error for no enforcer", func() {
					So(err, ShouldResemble, fmt.Errorf("Enforcer cannot be nil"))
				})
			})

			Convey("When I try to initiate the supervisor with enforcer", func() {
				var rpcwrperreq rpcwrapper.Request
				var rpcwrperres rpcwrapper.Response

				rpcwrperreq.HashAuth = []byte{0x47, 0xBE, 0x1A, 0x01, 0x47, 0x4F, 0x4A, 0x7A, 0xB5, 0xDA, 0x97, 0x46, 0xF3, 0x98, 0x50, 0x86, 0xB1, 0xF7, 0x05, 0x65, 0x6F, 0x58, 0x8C, 0x2C, 0x23, 0x9B, 0xA2, 0x82, 0x40, 0x45, 0x24, 0x45}
				rpcwrperreq.Payload = initTestSupReqPayload()
				rpcwrperres.Status = ""

				digest := hmac.New(sha256.New, []byte(os.Getenv("STATS_SECRET")))
				if _, err := digest.Write(structhash.Dump(rpcwrperreq.Payload, 1)); err != nil {
					So(err, ShouldBeNil)
				}
				rpcwrperreq.HashAuth = digest.Sum(nil)

				collector := &collector.DefaultCollector{}
				secret := secrets.NewPSKSecrets([]byte("Dummy Test Password"))
				server.Enforcer = enforcer.NewWithDefaults("someServerID", collector, nil, secret, constants.LocalContainer, "/proc").(*enforcer.Datapath)

				err := server.InitSupervisor(rpcwrperreq, &rpcwrperres)

				Convey("Then I should get error for no enforcer", func() {
					So(err, ShouldBeNil)
				})
			})
			os.Setenv("STATSCHANNEL_PATH", "")
			os.Setenv("STATS_SECRET", "")

		})
	})
}

func TestLaunchRemoteEnforcer(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	Convey("When I try to retrieve rpc server handle", t, func() {
		//rpcHdl := rpcwrapper.NewRPCServer()
		rpcHdl := mock_rpcwrapper.NewMockRPCServer(ctrl)

		Convey("Then rpcHdl should resemble rpcwrapper struct", func() {
			So(rpcHdl, ShouldNotBeNil)
		})

		Convey("When I try to create new server with no env set", func() {
			rpcHdl.EXPECT().StartServer(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			var service enforcer.PacketProcessor
			pcchan := "/tmp/test.sock"
			secret := "mysecret"
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get error for no stats", func() {
				So(server, ShouldBeNil)
				So(err, ShouldResemble, fmt.Errorf("No path to stats socket provided"))
			})
		})

		Convey("When I try to create new server with env set", func() {
			os.Setenv("STATSCHANNEL_PATH", "/tmp/test.sock")
			os.Setenv("STATS_SECRET", "mysecret")
			var service enforcer.PacketProcessor
			pcchan := os.Getenv("STATSCHANNEL_PATH")
			secret := os.Getenv("STATS_SECRET")
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get no error", func() {
				So(server, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("When I try to start the server", func() {
				os.Setenv("APORETO_ENV_SOCKET_PATH", "/tmp/test.sock")
				envpipe := os.Getenv(envSocketPath)
				rpcHdl.EXPECT().StartServer("unix", envpipe, server).Times(1).Return(nil)

				Convey("Then I expect to call start server one time with required parameters", func() {
					err := rpcHdl.StartServer("unix", envpipe, server)

					Convey("I should not get any error", func() {
						So(err, ShouldBeNil)
					})
				})
			})

			Convey("When I try to exit the enforcer", func() {
				server.statsclient = nil
				err := server.EnforcerExit(rpcwrapper.Request{}, &rpcwrapper.Response{})

				Convey("Then I should get no error", func() {
					So(err, ShouldBeNil)
				})
			})
			os.Setenv("STATSCHANNEL_PATH", "")
			os.Setenv("STATS_SECRET", "")
		})
	})
}

func TestSupervise(t *testing.T) {

	Convey("When I try to retrieve rpc server handle", t, func() {
		rpcHdl := rpcwrapper.NewRPCServer()

		Convey("Then rpcHdl should resemble rpcwrapper struct", func() {
			So(rpcHdl, ShouldNotBeNil)
		})

		Convey("When I try to create new server with no env set", func() {
			var service enforcer.PacketProcessor
			pcchan := "/tmp/test.sock"
			secret := "mysecret"
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get error for no stats", func() {
				So(server, ShouldBeNil)
				So(err, ShouldResemble, fmt.Errorf("No path to stats socket provided"))
			})
		})

		Convey("When I try to create new server with env set", func() {
			os.Setenv("STATSCHANNEL_PATH", "/tmp/test.sock")
			os.Setenv("STATS_SECRET", "zsGt6jhc1DkE0cHcv8HtJl_iP-8K_zPX4u0TUykDJSg=")
			var service enforcer.PacketProcessor
			pcchan := os.Getenv("STATSCHANNEL_PATH")
			secret := os.Getenv("STATS_SECRET")
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get no error", func() {
				So(server, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("When I try to send supervise command", func() {
				var rpcwrperreq rpcwrapper.Request
				var rpcwrperres rpcwrapper.Response

				rpcwrperreq.HashAuth = []byte{0x14, 0x5E, 0x0A, 0x3B, 0x50, 0xA3, 0xFF, 0xBC, 0xD5, 0x1B, 0x25, 0x21, 0x7D, 0x32, 0xD2, 0x02, 0x9F, 0x3A, 0xBE, 0xDC, 0x1F, 0xBB, 0xB7, 0x32, 0xFB, 0x91, 0x63, 0xA0, 0xF8, 0xE4, 0x43, 0x80}
				rpcwrperreq.Payload = initTestSupPayload()
				rpcwrperres.Status = ""

				digest := hmac.New(sha256.New, []byte(os.Getenv("STATS_SECRET")))
				if _, err := digest.Write(structhash.Dump(rpcwrperreq.Payload, 1)); err != nil {
					So(err, ShouldBeNil)
				}
				rpcwrperreq.HashAuth = digest.Sum(nil)

				c := &collector.DefaultCollector{}
				secrets := secrets.NewPSKSecrets([]byte("test password"))
				e := enforcer.NewWithDefaults("serverID", c, nil, secrets, constants.LocalContainer, "/proc")

				server.Supervisor, _ = supervisor.NewSupervisor(c, e, constants.LocalContainer, constants.IPTables, []string{})

				//err := server.Supervise(rpcwrperreq, &rpcwrperres)

				Convey("Then I should get no error", func() {
					//So(err, ShouldBeNil)
				})
			})

      Convey("When I try to send unsupervise command", func() {
        var rpcwrperreq rpcwrapper.Request
        var rpcwrperres rpcwrapper.Response

        rpcwrperreq.HashAuth = []byte{0x14, 0x5E, 0x0A, 0x3B, 0x50, 0xA3, 0xFF, 0xBC, 0xD5, 0x1B, 0x25, 0x21, 0x7D, 0x32, 0xD2, 0x02, 0x9F, 0x3A, 0xBE, 0xDC, 0x1F, 0xBB, 0xB7, 0x32, 0xFB, 0x91, 0x63, 0xA0, 0xF8, 0xE4, 0x43, 0x80}
        rpcwrperreq.Payload = initTestSupPayload()
        rpcwrperres.Status = ""

        digest := hmac.New(sha256.New, []byte(os.Getenv("STATS_SECRET")))
        if _, err := digest.Write(structhash.Dump(rpcwrperreq.Payload, 1)); err != nil {
          So(err, ShouldBeNil)
        }
        rpcwrperreq.HashAuth = digest.Sum(nil)

        c := &collector.DefaultCollector{}
        secrets := secrets.NewPSKSecrets([]byte("test password"))
        e := enforcer.NewWithDefaults("serverID", c, nil, secrets, constants.LocalContainer, "/proc")

        server.Supervisor, _ = supervisor.NewSupervisor(c, e, constants.LocalContainer, constants.IPTables, []string{})

        //err := server.Unsupervise(rpcwrperreq, &rpcwrperres)

        Convey("Then I should get no error", func() {
          //So(err, ShouldBeNil)
        })
      })
			os.Setenv("STATSCHANNEL_PATH", "")
			os.Setenv("STATS_SECRET", "")
		})
	})
}


func TestEnforce(t *testing.T) {

	Convey("When I try to retrieve rpc server handle", t, func() {
		rpcHdl := rpcwrapper.NewRPCServer()

		Convey("Then rpcHdl should resemble rpcwrapper struct", func() {
			So(rpcHdl, ShouldNotBeNil)
		})

		Convey("When I try to create new server with no env set", func() {
			var service enforcer.PacketProcessor
			pcchan := "/tmp/test.sock"
			secret := "mysecret"
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get error for no stats", func() {
				So(server, ShouldBeNil)
				So(err, ShouldResemble, fmt.Errorf("No path to stats socket provided"))
			})
		})

		Convey("When I try to create new server with env set", func() {
			os.Setenv("STATSCHANNEL_PATH", "/tmp/test.sock")
			os.Setenv("STATS_SECRET", "KMvm4a6kgLLma5NitOMGx2f9k21G3nrAaLbgA5zNNHM=")
			var service enforcer.PacketProcessor
			pcchan := os.Getenv("STATSCHANNEL_PATH")
			secret := os.Getenv("STATS_SECRET")
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get no error", func() {
				So(server, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("When I try to send enforce command", func() {
				var rpcwrperreq rpcwrapper.Request
				var rpcwrperres rpcwrapper.Response

				rpcwrperreq.HashAuth = []byte{0xDE,0xBD,0x1C,0x6A,0x2A,0x51,0xC0,0x02,0x4B,0xD7,0xD1,0x82,0x78,0x8A,0xC4,0xF1,0xBE,0xBF,0x00,0x89,0x47,0x0F,0x13,0x71,0xAB,0x4C,0x0D,0xD9,0x9D,0x85,0x45,0x04}
				rpcwrperreq.Payload = initTestEnfPayload()
				rpcwrperres.Status = ""

				digest := hmac.New(sha256.New, []byte(os.Getenv("STATS_SECRET")))
				if _, err := digest.Write(structhash.Dump(rpcwrperreq.Payload, 1)); err != nil {
					So(err, ShouldBeNil)
				}
				rpcwrperreq.HashAuth = digest.Sum(nil)

        collector := &collector.DefaultCollector{}
				secret := secrets.NewPSKSecrets([]byte("Dummy Test Password"))
				server.Enforcer = enforcer.NewWithDefaults("someServerID", collector, nil, secret, constants.LocalServer, "/proc").(*enforcer.Datapath)

				err := server.Enforce(rpcwrperreq, &rpcwrperres)

				Convey("Then I should get no error", func() {
					So(err, ShouldBeNil)
				})
			})
			os.Setenv("STATSCHANNEL_PATH", "")
			os.Setenv("STATS_SECRET", "")
		})
	})
}


func TestUnEnforce(t *testing.T) {

	Convey("When I try to retrieve rpc server handle", t, func() {
		rpcHdl := rpcwrapper.NewRPCServer()

		Convey("Then rpcHdl should resemble rpcwrapper struct", func() {
			So(rpcHdl, ShouldNotBeNil)
		})

		Convey("When I try to create new server with no env set", func() {
			var service enforcer.PacketProcessor
			pcchan := "/tmp/test.sock"
			secret := "mysecret"
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get error for no stats", func() {
				So(server, ShouldBeNil)
				So(err, ShouldResemble, fmt.Errorf("No path to stats socket provided"))
			})
		})

		Convey("When I try to create new server with env set", func() {
			os.Setenv("STATSCHANNEL_PATH", "/tmp/test.sock")
			os.Setenv("STATS_SECRET", "KMvm4a6kgLLma5NitOMGx2f9k21G3nrAaLbgA5zNNHM=")
			var service enforcer.PacketProcessor
			pcchan := os.Getenv("STATSCHANNEL_PATH")
			secret := os.Getenv("STATS_SECRET")
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get no error", func() {
				So(server, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("When I try to send Unenforce command without enforcer", func() {
				var rpcwrperreq rpcwrapper.Request
				var rpcwrperres rpcwrapper.Response

				rpcwrperreq.HashAuth = []byte{0xDE,0xBD,0x1C,0x6A,0x2A,0x51,0xC0,0x02,0x4B,0xD7,0xD1,0x82,0x78,0x8A,0xC4,0xF1,0xBE,0xBF,0x00,0x89,0x47,0x0F,0x13,0x71,0xAB,0x4C,0x0D,0xD9,0x9D,0x85,0x45,0x04}
				rpcwrperreq.Payload = initTestUnEnfPayload()
				rpcwrperres.Status = ""

				digest := hmac.New(sha256.New, []byte(os.Getenv("STATS_SECRET")))
				if _, err := digest.Write(structhash.Dump(rpcwrperreq.Payload, 1)); err != nil {
					So(err, ShouldBeNil)
				}
				rpcwrperreq.HashAuth = digest.Sum(nil)

        collector := &collector.DefaultCollector{}
				secret := secrets.NewPSKSecrets([]byte("Dummy Test Password"))
				server.Enforcer = enforcer.NewWithDefaults("b06f47830f64", collector, nil, secret, constants.LocalContainer, "/proc").(*enforcer.Datapath)

				err := server.Unenforce(rpcwrperreq, &rpcwrperres)

				Convey("Then I should get error", func() {
					So(err, ShouldResemble,fmt.Errorf("ContextID not found in Enforcer"))
				})
			})
			os.Setenv("STATSCHANNEL_PATH", "")
			os.Setenv("STATS_SECRET", "")
		})
	})
}


func TestUnSupervise(t *testing.T) {

	Convey("When I try to retrieve rpc server handle", t, func() {
		rpcHdl := rpcwrapper.NewRPCServer()

		Convey("Then rpcHdl should resemble rpcwrapper struct", func() {
			So(rpcHdl, ShouldNotBeNil)
		})

		Convey("When I try to create new server with no env set", func() {
			var service enforcer.PacketProcessor
			pcchan := "/tmp/test.sock"
			secret := "mysecret"
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get error for no stats", func() {
				So(server, ShouldBeNil)
				So(err, ShouldResemble, fmt.Errorf("No path to stats socket provided"))
			})
		})

		Convey("When I try to create new server with env set", func() {
			os.Setenv("STATSCHANNEL_PATH", "/tmp/test.sock")
			os.Setenv("STATS_SECRET", "zsGt6jhc1DkE0cHcv8HtJl_iP-8K_zPX4u0TUykDJSg=")
			var service enforcer.PacketProcessor
			pcchan := os.Getenv("STATSCHANNEL_PATH")
			secret := os.Getenv("STATS_SECRET")
			server, err := NewServer(service, rpcHdl, pcchan, secret)

			Convey("Then I should get no error", func() {
				So(server, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("When I try to send unsupervise command before supervise", func() {
				var rpcwrperreq rpcwrapper.Request
				var rpcwrperres rpcwrapper.Response

				rpcwrperreq.HashAuth = []byte{0x14, 0x5E, 0x0A, 0x3B, 0x50, 0xA3, 0xFF, 0xBC, 0xD5, 0x1B, 0x25, 0x21, 0x7D, 0x32, 0xD2, 0x02, 0x9F, 0x3A, 0xBE, 0xDC, 0x1F, 0xBB, 0xB7, 0x32, 0xFB, 0x91, 0x63, 0xA0, 0xF8, 0xE4, 0x43, 0x80}
				rpcwrperreq.Payload = initTestUnSupPayload()
				rpcwrperres.Status = ""

				digest := hmac.New(sha256.New, []byte(os.Getenv("STATS_SECRET")))
				if _, err := digest.Write(structhash.Dump(rpcwrperreq.Payload, 1)); err != nil {
					So(err, ShouldBeNil)
				}
				rpcwrperreq.HashAuth = digest.Sum(nil)

				c := &collector.DefaultCollector{}
				secrets := secrets.NewPSKSecrets([]byte("test password"))
				e := enforcer.NewWithDefaults("ac0d3577e808", c, nil, secrets, constants.LocalContainer, "/proc")

				server.Supervisor, _ = supervisor.NewSupervisor(c, e, constants.LocalContainer, constants.IPTables, []string{})

				err := server.Unsupervise(rpcwrperreq, &rpcwrperres)

				Convey("Then I should get no error", func() {
					So(err, ShouldResemble, fmt.Errorf("Cannot find policy version"))
				})
			})
			os.Setenv("STATSCHANNEL_PATH", "")
			os.Setenv("STATS_SECRET", "")
		})
	})
}
