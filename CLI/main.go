package main

import (
	cutil "Archcopy/util"
	"ArchcopyCLI/util"
	"fmt"
	"os"

	"github.com/integrii/flaggy"
)

const Debug bool = false

type UIRPC struct {
	// Common
	PSK           string
	SSHPubKey     string
	SSHPubKeyPass string
	CACertificate string
	// Write
	OutputURL string
	InputURL  string
	// Client
	ClientCertificate    string
	ClientCertificateKey string
	// Server
	Auto                 bool
	ServerCertificate    string
	ServerCertificateKey string
	SlaveURL             string
	// SSH
	SingleSession bool
}

type UIStruct struct {
	RPC                  UIRPC
	Resume               bool
	Sparse               bool
	HaveRoot             bool
	Verify               bool
	DryRun               bool
	ContinueOnError      bool
	AllowOverwrite       bool
	InputRemainder       bool
	InputFile            []string
	OutputFile           string
	InputDirectory       string
	OutputDirectory      string
	PreserveRelativePath bool
	ShowStackTrace       bool
}

var UI UIStruct

func main() {
	fmt.Print("Archcopy 1. THIS IS EXPERIMENTAL SOFTWARE. DO NOT USE IT FOR ANYTHING IMPORTANT.\n")

	flaggy.SetVersion("1")

	flaggy.StringSlice(&UI.InputFile, "if", "inputfile", "Source file")
	flaggy.Bool(&UI.InputRemainder, "ic", "inputconventional", "Treat all command line parameters after -- as input filenames. An output directory must be specified with -od.")
	flaggy.String(&UI.InputDirectory, "id", "inputdirectory", "Source directory")
	flaggy.String(&UI.OutputFile, "of", "outputfile", "Destination filename")
	flaggy.String(&UI.OutputDirectory, "od", "outputdirectory", "Destination directory")
	flaggy.Bool(&UI.PreserveRelativePath, "p", "preserverelativepath", "Create source relative path in destination folder.")
	flaggy.Bool(&UI.AllowOverwrite, "f", "force", "Overwrite existing destination files.")
	flaggy.Bool(&UI.Resume, "r", "resume", "Resume an interrupted transfer.")
	flaggy.Bool(&UI.Sparse, "s", "sparse", "Write 4K blocks containing only zero bytes as sparse extents.")
	flaggy.Bool(&UI.Verify, "v", "verify", "Verify that the hashes of the source and destination files are identical.")
	flaggy.Bool(&UI.ContinueOnError, "c", "continue", "Continue with other files in the event of errors. Default behavior is to exit on any error.")
	flaggy.Bool(&UI.DryRun, "d", "dryrun", "Dry run only. List files to be transferred, but do not transfer them.")

	flaggy.String(&UI.RPC.OutputURL, "ro", "remoteoutputurl", "RPC: URL for remote output.")
	flaggy.String(&UI.RPC.InputURL, "ri", "remoteinputurl", "RPC: URL for remote input.")
	flaggy.String(&UI.RPC.PSK, "rp", "rpcpsk", "RPC: Pre-shared key.")
	flaggy.String(&UI.RPC.SSHPubKey, "rk", "rpcsshkey", "RPC: SSH key pair (default: ~/.ssh/id_rsa)")
	flaggy.String(&UI.RPC.CACertificate, "ca", "rpccacert", "RPC: CA certificate.")
	flaggy.String(&UI.RPC.ClientCertificate, "cc", "rpccert", "RPC: Client certificate.")
	flaggy.String(&UI.RPC.ClientCertificateKey, "ck", "rpccertkey", "RPC: Client certificate key.")

	// SelftestSubcommand := flaggy.NewSubcommand("SelfTest")
	// SelftestSubcommand.Description = "Perform internal self-checks. Intended for debugging only."
	// flaggy.AttachSubcommand(SelftestSubcommand, 1)

	GenTLSSubcommand := flaggy.NewSubcommand("GenerateCerts")
	GenTLSSubcommand.Description = "Generate certificates for use with Archcopy TLS."
	flaggy.AttachSubcommand(GenTLSSubcommand, 1)

	SlaveSubcommand := flaggy.NewSubcommand("slave")
	SlaveSubcommand.Description = "Listen on specified address for remote connections."
	SlaveSubcommand.String(&UI.RPC.SlaveURL, "sl", "listenurl", "URL to listen for connections.")
	SlaveSubcommand.Bool(&UI.RPC.Auto, "sa", "auto", "Generate PSK automatically.")
	SlaveSubcommand.Bool(&UI.RPC.SingleSession, "st", "singlesession", "Terminate after client disconnects.")
	SlaveSubcommand.String(&UI.RPC.ServerCertificate, "sc", "certificate", "TLS certificate")
	SlaveSubcommand.String(&UI.RPC.ServerCertificateKey, "sk", "certificatekey", "TLS certificate private key")
	flaggy.AttachSubcommand(SlaveSubcommand, 1)

	BlakeSubcommand := flaggy.NewSubcommand("blake2bhash")
	flaggy.AttachSubcommand(BlakeSubcommand, 1)

	flaggy.Parse()

	UI.ShowStackTrace = true

	util.StackTrace = UI.ShowStackTrace
	cutil.StackTrace = UI.ShowStackTrace

	var rc = 0
	if GenTLSSubcommand.Used {
		rc = MakeCerts()
		// } else if SelftestSubcommand.Used {
		// 	rc = Test()
	} else if SlaveSubcommand.Used {
		rc = Slave()
	} else if BlakeSubcommand.Used {
		// Workaround for a serious flaggy bug that the maintainer won't acknowledge.
		// https://github.com/integrii/flaggy/issues/65

		L := len(UI.InputFile)
		if L > 1 {
			UI.InputFile = UI.InputFile[:L/2]
		}
		if UI.InputRemainder {
			UI.InputFile = append(UI.InputFile, flaggy.TrailingArguments...)
		}
		rc = Blake2bHash()
	} else {
		if UI.InputRemainder {
			UI.InputFile = append(UI.InputFile, flaggy.TrailingArguments...)
		}
		rc = Copy()
	}
	os.Exit(rc)
}
