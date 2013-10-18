package sampling

import (
	"bytes"
	"strconv"
)

var interface_types = []string{"",
	"other(1)", // none of the following
	"regular1822(2)",
	"hdh1822(3)",
	"ddnX25(4)",
	"rfc877x25(5)",
	"ethernetCsmacd(6)", // for all ethernet-like interfaces,
	// regardless of speed, as per RFC3635
	"iso88023Csmacd(7)", // Deprecated via RFC3635
	// ethernetCsmacd (6) should be used instead
	"iso88024TokenBus(8)",
	"iso88025TokenRing(9)",
	"iso88026Man(10)",
	"starLan(11)", // Deprecated via RFC3635
	// ethernetCsmacd (6) should be used instead
	"proteon10Mbit(12)",
	"proteon80Mbit(13)",
	"hyperchannel(14)",
	"fddi(15)",
	"lapb(16)",
	"sdlc(17)",
	"ds1(18)",       // DS1-MIB
	"e1(19)",        // Obsolete see DS1-MIB
	"basicISDN(20)", // no longer used
	// see also RFC2127
	"primaryISDN(21)", // no longer used
	// see also RFC2127
	"propPointToPointSerial(22)", // proprietary serial
	"ppp(23)",
	"softwareLoopback(24)",
	"eon(25)", // CLNP over IP
	"ethernet3Mbit(26)",
	"nsip(27)",       // XNS over IP
	"slip(28)",       // generic SLIP
	"ultra(29)",      // ULTRA technologies
	"ds3(30)",        // DS3-MIB
	"sip(31)",        // SMDS, coffee
	"frameRelay(32)", // DTE only.
	"rs232(33)",
	"para(34)",       // parallel-port
	"arcnet(35)",     // arcnet
	"arcnetPlus(36)", // arcnet plus
	"atm(37)",        // ATM cells
	"miox25(38)",
	"sonet(39)", // SONET or SDH
	"x25ple(40)",
	"iso88022llc(41)",
	"localTalk(42)",
	"smdsDxi(43)",
	"frameRelayService(44)", // FRNETSERV-MIB
	"v35(45)",
	"hssi(46)",
	"hippi(47)",
	"modem(48)", // Generic modem
	"aal5(49)",  // AAL5 over ATM
	"sonetPath(50)",
	"sonetVT(51)",
	"smdsIcip(52)",               // SMDS InterCarrier Interface
	"propVirtual(53)",            // proprietary virtual/internal
	"propMultiplexor(54)",        // proprietary multiplexing
	"ieee80212(55)",              // 100BaseVG
	"fibreChannel(56)",           // Fibre Channel
	"hippiInterface(57)",         // HIPPI interfaces
	"frameRelayInterconnect(58)", // Obsolete, use either
	// frameRelay(32) or
	// frameRelayService(44).
	"aflane8023(59)", // ATM Emulated LAN for 802.3
	"aflane8025(60)", // ATM Emulated LAN for 802.5
	"cctEmul(61)",    // ATM Emulated circuit
	"fastEther(62)",  // Obsoleted via RFC3635
	// ethernetCsmacd (6) should be used instead
	"isdn(63)",        // ISDN and X.25
	"v11(64)",         // CCITT V.11/X.21
	"v36(65)",         // CCITT V.36
	"g703at64k(66)",   // CCITT G703 at 64Kbps
	"g703at2mb(67)",   // Obsolete see DS1-MIB
	"qllc(68)",        // SNA QLLC                   "
	"fastEtherFX(69)", // Obsoleted via RFC3635
	// ethernetCsmacd (6) should be used instead
	"channel(70)",             // channel                   "
	"ieee80211(71)",           // radio spread spectrum
	"ibm370parChan(72)",       // IBM System 360/370 OEMI Channel
	"escon(73)",               // IBM Enterprise Systems Connection
	"dlsw(74)",                // Data Link Switching
	"isdns(75)",               // ISDN S/T interface
	"isdnu(76)",               // ISDN U interface
	"lapd(77)",                // Link Access Protocol D
	"ipSwitch(78)",            // IP Switching Objects
	"rsrb(79)",                // Remote Source Route Bridging
	"atmLogical(80)",          // ATM Logical Port
	"ds0(81)",                 // Digital Signal Level 0
	"ds0Bundle(82)",           // group of ds0s on the same ds1
	"bsc(83)",                 // Bisynchronous Protocol
	"async(84)",               // Asynchronous Protocol
	"cnr(85)",                 // Combat Net Radio
	"iso88025Dtr(86)",         // ISO 802.5r DTR
	"eplrs(87)",               // Ext Pos Loc Report Sys
	"arap(88)",                // Appletalk Remote Access Protocol
	"propCnls(89)",            // Proprietary Connectionless Protocol
	"hostPad(90)",             // CCITT-ITU X.29 PAD Protocol
	"termPad(91)",             // CCITT-ITU X.3 PAD Facility
	"frameRelayMPI(92)",       // Multiproto Interconnect over FR
	"x213(93)",                // CCITT-ITU X213
	"adsl(94)",                // Asymmetric Digital Subscriber Loop
	"radsl(95)",               // Rate-Adapt. Digital Subscriber Loop
	"sdsl(96)",                // Symmetric Digital Subscriber Loop
	"vdsl(97)",                // Very H-Speed Digital Subscrib. Loop
	"iso88025CRFPInt(98)",     // ISO 802.5 CRFP
	"myrinet(99)",             // Myricom Myrinet
	"voiceEM(100)",            // voice recEive and transMit
	"voiceFXO(101)",           // voice Foreign Exchange Office
	"voiceFXS(102)",           // voice Foreign Exchange Station
	"voiceEncap(103)",         // voice encapsulation
	"voiceOverIp(104)",        // voice over IP encapsulation
	"atmDxi(105)",             // ATM DXI
	"atmFuni(106)",            // ATM FUNI
	"atmIma (107)",            // ATM IMA
	"pppMultilinkBundle(108)", // PPP Multilink Bundle
	"ipOverCdlc (109)",        // IBM ipOverCdlc
	"ipOverClaw (110)",        // IBM Common Link Access to Workstn
	"stackToStack (111)",      // IBM stackToStack
	"virtualIpAddress (112)",  // IBM VIPA
	"mpc (113)",               // IBM multi-protocol channel support
	"ipOverAtm (114)",         // IBM ipOverAtm
	"iso88025Fiber (115)",     // ISO 802.5j Fiber Token Ring
	"tdlc (116)",              // IBM twinaxial data link control
	"gigabitEthernet (117)",   // Obsoleted via RFC3635
	// ethernetCsmacd (6) should be used instead
	"hdlc (118)",                       // HDLC
	"lapf (119)",                       // LAP F
	"v37 (120)",                        // V.37
	"x25mlp (121)",                     // Multi-Link Protocol
	"x25huntGroup (122)",               // X25 Hunt Group
	"transpHdlc (123)",                 // Transp HDLC
	"interleave (124)",                 // Interleave channel
	"fast (125)",                       // Fast channel
	"ip (126)",                         // IP (for APPN HPR in IP networks)
	"docsCableMaclayer (127)",          // CATV Mac Layer
	"docsCableDownstream (128)",        // CATV Downstream interface
	"docsCableUpstream (129)",          // CATV Upstream interface
	"a12MppSwitch (130)",               // Avalon Parallel Processor
	"tunnel (131)",                     // Encapsulation interface
	"coffee (132)",                     // coffee pot
	"ces (133)",                        // Circuit Emulation Service
	"atmSubInterface (134)",            // ATM Sub Interface
	"l2vlan (135)",                     // Layer 2 Virtual LAN using 802.1Q
	"l3ipvlan (136)",                   // Layer 3 Virtual LAN using IP
	"l3ipxvlan (137)",                  // Layer 3 Virtual LAN using IPX
	"digitalPowerline (138)",           // IP over Power Lines
	"mediaMailOverIp (139)",            // Multimedia Mail over IP
	"dtm (140)",                        // Dynamic syncronous Transfer Mode
	"dcn (141)",                        // Data Communications Network
	"ipForward (142)",                  // IP Forwarding Interface
	"msdsl (143)",                      // Multi-rate Symmetric DSL
	"ieee1394 (144)",                   // IEEE1394 High Performance Serial Bus
	"if-gsn (145)",                     //   HIPPI-6400
	"dvbRccMacLayer (146)",             // DVB-RCC MAC Layer
	"dvbRccDownstream (147)",           // DVB-RCC Downstream Channel
	"dvbRccUpstream (148)",             // DVB-RCC Upstream Channel
	"atmVirtual (149)",                 // ATM Virtual Interface
	"mplsTunnel (150)",                 // MPLS Tunnel Virtual Interface
	"srp (151)",                        // Spatial Reuse Protocol
	"voiceOverAtm (152)",               // Voice Over ATM
	"voiceOverFrameRelay (153)",        // Voice Over Frame Relay
	"idsl (154)",                       // Digital Subscriber Loop over ISDN
	"compositeLink (155)",              // Avici Composite Link Interface
	"ss7SigLink (156)",                 // SS7 Signaling Link
	"propWirelessP2P (157)",            //  Prop. P2P wireless interface
	"frForward (158)",                  // Frame Forward Interface
	"rfc1483 (159)",                    // Multiprotocol over ATM AAL5
	"usb (160)",                        // USB Interface
	"ieee8023adLag (161)",              // IEEE 802.3ad Link Aggregate
	"bgppolicyaccounting (162)",        // BGP Policy Accounting
	"frf16MfrBundle (163)",             // FRF .16 Multilink Frame Relay
	"h323Gatekeeper (164)",             // H323 Gatekeeper
	"h323Proxy (165)",                  // H323 Voice and Video Proxy
	"mpls (166)",                       // MPLS                   "
	"mfSigLink (167)",                  // Multi-frequency signaling link
	"hdsl2 (168)",                      // High Bit-Rate DSL - 2nd generation
	"shdsl (169)",                      // Multirate HDSL2
	"ds1FDL (170)",                     // Facility Data Link 4Kbps on a DS1
	"pos (171)",                        // Packet over SONET/SDH Interface
	"dvbAsiIn (172)",                   // DVB-ASI Input
	"dvbAsiOut (173)",                  // DVB-ASI Output
	"plc (174)",                        // Power Line Communtications
	"nfas (175)",                       // Non Facility Associated Signaling
	"tr008 (176)",                      // TR008
	"gr303RDT (177)",                   // Remote Digital Terminal
	"gr303IDT (178)",                   // Integrated Digital Terminal
	"isup (179)",                       // ISUP
	"propDocsWirelessMaclayer (180)",   // Cisco proprietary Maclayer
	"propDocsWirelessDownstream (181)", // Cisco proprietary Downstream
	"propDocsWirelessUpstream (182)",   // Cisco proprietary Upstream
	"hiperlan2 (183)",                  // HIPERLAN Type 2 Radio Interface
	"propBWAp2Mp (184)",                // PropBroadbandWirelessAccesspt2multipt
	// use of this iftype for IEEE 802.16 WMAN
	// interfaces as per IEEE Std 802.16f is
	// deprecated and ifType 237 should be used instead.
	"sonetOverheadChannel (185)",          // SONET Overhead Channel
	"digitalWrapperOverheadChannel (186)", // Digital Wrapper
	"aal2 (187)",                          // ATM adaptation layer 2
	"radioMAC (188)",                      // MAC layer over radio links
	"atmRadio (189)",                      // ATM over radio links
	"imt (190)",                           // Inter Machine Trunks
	"mvl (191)",                           // Multiple Virtual Lines DSL
	"reachDSL (192)",                      // Long Reach DSL
	"frDlciEndPt (193)",                   // Frame Relay DLCI End Point
	"atmVciEndPt (194)",                   // ATM VCI End Point
	"opticalChannel (195)",                // Optical Channel
	"opticalTransport (196)",              // Optical Transport
	"propAtm (197)",                       //  Proprietary ATM
	"voiceOverCable (198)",                // Voice Over Cable Interface
	"infiniband (199)",                    // Infiniband
	"teLink (200)",                        // TE Link
	"q2931 (201)",                         // Q.2931
	"virtualTg (202)",                     // Virtual Trunk Group
	"sipTg (203)",                         // SIP Trunk Group
	"sipSig (204)",                        // SIP Signaling
	"docsCableUpstreamChannel (205)",      // CATV Upstream Channel
	"econet (206)",                        // Acorn Econet
	"pon155 (207)",                        // FSAN 155Mb Symetrical PON interface
	"pon622 (208)",                        // FSAN622Mb Symetrical PON interface
	"bridge (209)",                        // Transparent bridge interface
	"linegroup (210)",                     // Interface common to multiple lines
	"voiceEMFGD (211)",                    // voice E&M Feature Group D
	"voiceFGDEANA (212)",                  // voice FGD Exchange Access North American
	"voiceDID (213)",                      // voice Direct Inward Dialing
	"mpegTransport (214)",                 // MPEG transport interface
	"sixToFour (215)",                     // 6to4 interface (DEPRECATED)
	"gtp (216)",                           // GTP (GPRS Tunneling Protocol)
	"pdnEtherLoop1 (217)",                 // Paradyne EtherLoop 1
	"pdnEtherLoop2 (218)",                 // Paradyne EtherLoop 2
	"opticalChannelGroup (219)",           // Optical Channel Group
	"homepna (220)",                       // HomePNA ITU-T G.989
	"gfp (221)",                           // Generic Framing Procedure (GFP)
	"ciscoISLvlan (222)",                  // Layer 2 Virtual LAN using Cisco ISL
	"actelisMetaLOOP (223)",               // Acteleis proprietary MetaLOOP High Speed Link
	"fcipLink (224)",                      // FCIP Link
	"rpr (225)",                           // Resilient Packet Ring Interface Type
	"qam (226)",                           // RF Qam Interface
	"lmp (227)",                           // Link Management Protocol
	"cblVectaStar (228)",                  // Cambridge Broadband Networks Limited VectaStar
	"docsCableMCmtsDownstream (229)",      // CATV Modular CMTS Downstream Interface
	"adsl2 (230)",                         // Asymmetric Digital Subscriber Loop Version 2
	// (DEPRECATED/OBSOLETED - please use adsl2plus 238 instead)
	"macSecControlledIF (231)",   // MACSecControlled
	"macSecUncontrolledIF (232)", // MACSecUncontrolled
	"aviciOpticalEther (233)",    // Avici Optical Ethernet Aggregate
	"atmbond (234)",              // atmbond
	"voiceFGDOS (235)",           // voice FGD Operator Services
	"mocaVersion1 (236)",         // MultiMedia over Coax Alliance (MoCA) Interface
	// as documented in information provided privately to IANA
	"ieee80216WMAN (237)", // IEEE 802.16 WMAN interface
	"adsl2plus (238)",     // Asymmetric Digital Subscriber Loop Version 2,
	// Version 2 Plus and all variants
	"dvbRcsMacLayer (239)",          // DVB-RCS MAC Layer
	"dvbTdm (240)",                  // DVB Satellite TDM
	"dvbRcsTdma (241)",              // DVB-RCS TDMA
	"x86Laps (242)",                 // LAPS based on ITU-T X.86/Y.1323
	"wwanPP (243)",                  // 3GPP WWAN
	"wwanPP2 (244)",                 // 3GPP2 WWAN
	"voiceEBS (245)",                // voice P-phone EBS physical interface
	"ifPwType (246)",                // Pseudowire interface type
	"ilan (247)",                    // Internal LAN on a bridge per IEEE 802.1ap
	"pip (248)",                     // Provider Instance Port on a bridge per IEEE 802.1ah PBB
	"aluELP (249)",                  // Alcatel-Lucent Ethernet Link Protection
	"gpon (250)",                    // Gigabit-capable passive optical networks (G-PON) as per ITU-T G.948
	"vdsl2 (251)",                   // Very high speed digital subscriber line Version 2 (as per ITU-T Recommendation G.993.2)
	"capwapDot11Profile (252)",      // WLAN Profile Interface
	"capwapDot11Bss (253)",          // WLAN BSS Interface
	"capwapWtpVirtualRadio (254)",   // WTP Virtual Radio Interface
	"bits (255)",                    // bitsport
	"docsCableUpstreamRfPort (256)", // DOCSIS CATV Upstream RF Port
	"cableDownstreamRfPort (257)",   // CATV downstream RF port
	"vmwareVirtualNic (258)",        // VMware Virtual Network Interface
	"ieee802154 (259)",              // IEEE 802.15.4 WPAN interface
	"otnOdu (260)",                  // OTN Optical Data Unit
	"otnOtu (261)",                  // OTN Optical channel Transport Unit
	"ifVfiType (262)",               // VPLS Forwarding Instance Interface Type
	"g9981 (263)",                   // G.998.1 bonded interface
	"g9982 (264)",                   // G.998.2 bonded interface
	"g9983 (265)",                   // G.998.3 bonded interface
	"aluEpon (266)",                 // Ethernet Passive Optical Networks (E-PON)
	"aluEponOnu (267)",              // EPON Optical Network Unit
	"aluEponPhysicalUni (268)",      // EPON physical User to Network interface
	"aluEponLogicalLink (269)",      // The emulation of a point-to-point link over the EPON layer
	"aluGponOnu (270)",              // GPON Optical Network Unit
	"aluGponPhysicalUni (271)",      // GPON physical User to Network interface
	"vmwareNicTeam (272)"}           // VMware NIC Team

var interface_status_list = []string{"",
	"up",
	"down",
	"testing",
	"unknown",
	"dormant",
	"notPresent",
	"lowerLayerDown"}

func interfaceType(ifType int32) string {
	if ifType > 0 && ifType < int32(len(interface_types)) {
		return interface_types[ifType]
	}
	return "unkown(" + strconv.FormatInt(int64(ifType), 10) + ")"
}
func statusString(ifAdminStatus, ifOperStatus int32) string {
	if ifAdminStatus == 1 && ifOperStatus == 1 {
		return "UP"
	} else {
		return "DOWN"
	}
}
func calcStatus(ifAdminStatus, ifOperStatus int32) int {
	if ifAdminStatus == 1 && ifOperStatus == 1 {
		return 1
	} else {
		return 0
	}
}

func interfaceStatusString(status int32) string {
	if status > 0 && status < int32(len(interface_status_list)) {
		return interface_status_list[status]
	}
	return "unkown(" + strconv.FormatInt(int64(status), 10) + ")"
}

func readInterface(params MContext, res map[string]interface{}) map[string]interface{} {
	new_row := map[string]interface{}{}
	new_row["ifIndex"] = GetInt32WithDefault(params, res, "1", -1)
	new_row["ifBit"] = 32
	new_row["ifDescr"] = GetStringWithDefault(params, res, "2")

	ifType := GetInt32WithDefault(params, res, "3", -1)
	new_row["ifType"] = ifType
	new_row["ifType__descr"] = interfaceType(ifType)

	new_row["ifMtu"] = GetInt32WithDefault(params, res, "4", -1)
	new_row["ifSpeed"] = GetUint64WithDefault(params, res, "5", 0)

	ifAdminStatus := GetInt32WithDefault(params, res, "7", -1)
	ifOpStatus := GetInt32WithDefault(params, res, "8", -1)
	new_row["ifAdminStatus"] = ifAdminStatus
	new_row["ifAdminStatus__descr"] = interfaceStatusString(ifAdminStatus)
	new_row["ifOpStatus"] = ifOpStatus
	new_row["ifOpStatus__descr"] = interfaceStatusString(ifOpStatus)
	new_row["ifStatus"] = calcStatus(ifAdminStatus, ifOpStatus)
	new_row["ifStatus__descr"] = statusString(ifAdminStatus, ifOpStatus)

	new_row["ifOpStatus"] = GetInt32WithDefault(params, res, "8", -1)
	new_row["ifLastChange"] = GetInt32WithDefault(params, res, "9", -1)
	new_row["ifInOctets"] = GetUint64WithDefault(params, res, "10", 0)
	new_row["ifInUcastPkts"] = GetUint64WithDefault(params, res, "11", 0)
	new_row["ifInNUcastPkts"] = GetUint64WithDefault(params, res, "12", 0)
	new_row["ifInDiscards"] = GetUint64WithDefault(params, res, "13", 0)
	new_row["ifInErrors"] = GetUint64WithDefault(params, res, "14", 0)
	new_row["ifInUnknownProtos"] = GetUint64WithDefault(params, res, "15", 0)
	new_row["ifOutOctets"] = GetUint64WithDefault(params, res, "16", 0)
	new_row["ifOutUcastPkts"] = GetUint64WithDefault(params, res, "17", 0)
	new_row["ifOutNUcastPkts"] = GetUint64WithDefault(params, res, "18", 0)
	new_row["ifOutDiscards"] = GetUint64WithDefault(params, res, "19", 0)
	new_row["ifOutErrors"] = GetUint64WithDefault(params, res, "20", 0)
	new_row["ifOutQLen"] = GetUint64WithDefault(params, res, "21", 0)
	new_row["ifSpecific"] = GetOidWithDefault(params, res, "22")
	return new_row
}

type portAll struct {
	snmpBase
}

func (self *portAll) Call(params MContext) (interface{}, error) {
	ifIndex, e := params.GetString("@ifIndex")
	if nil != e || 0 == len(ifIndex) {
		return nil, IsRequired("ifIndex")
	}

	oidBuffer := bytes.NewBuffer(make([]byte, 0, 900))
	for _, s := range []string{"1.3.6.1.2.1.2.2.1.2.",
		"1.3.6.1.2.1.2.2.1.3.",
		"1.3.6.1.2.1.2.2.1.4.",
		"1.3.6.1.2.1.2.2.1.5.",
		"1.3.6.1.2.1.2.2.1.6.",
		"1.3.6.1.2.1.2.2.1.7.",
		"1.3.6.1.2.1.2.2.1.8.",
		"1.3.6.1.2.1.2.2.1.9.",
		"1.3.6.1.2.1.2.2.1.10.",
		"1.3.6.1.2.1.2.2.1.11.",
		"1.3.6.1.2.1.2.2.1.12.",
		"1.3.6.1.2.1.2.2.1.13.",
		"1.3.6.1.2.1.2.2.1.14.",
		"1.3.6.1.2.1.2.2.1.15.",
		"1.3.6.1.2.1.2.2.1.16.",
		"1.3.6.1.2.1.2.2.1.17.",
		"1.3.6.1.2.1.2.2.1.18.",
		"1.3.6.1.2.1.2.2.1.19.",
		"1.3.6.1.2.1.2.2.1.20.",
		"1.3.6.1.2.1.2.2.1.21.",
		"1.3.6.1.2.1.2.2.1.22."} {
		oidBuffer.WriteString(s)
		oidBuffer.WriteString(ifIndex)
		oidBuffer.WriteString(",")
	}
	oidBuffer.Truncate(oidBuffer.Len() - 1)
	res, e := self.Get(params, oidBuffer.String())
	if nil != e {
		return nil, e
	}

	new_row := map[string]interface{}{}
	new_row["ifIndex"] = ifIndex
	new_row["ifBit"] = 32
	new_row["ifDescr"] = GetStringWithDefault(params, res, "1.3.6.1.2.1.2.2.1.2."+ifIndex)
	ifType := GetInt32WithDefault(params, res, "1.3.6.1.2.1.2.2.1.3."+ifIndex, -1)
	new_row["ifType"] = ifType
	new_row["ifType__descr"] = interfaceType(ifType)
	new_row["ifMtu"] = GetInt32WithDefault(params, res, "1.3.6.1.2.1.2.2.1.4."+ifIndex, -1)
	new_row["ifSpeed"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.5."+ifIndex, 0)

	ifAdminStatus := GetInt32WithDefault(params, res, "1.3.6.1.2.1.2.2.1.7."+ifIndex, -1)
	ifOpStatus := GetInt32WithDefault(params, res, "1.3.6.1.2.1.2.2.1.8."+ifIndex, -1)
	new_row["ifAdminStatus"] = ifAdminStatus
	new_row["ifAdminStatus__descr"] = interfaceStatusString(ifAdminStatus)
	new_row["ifOpStatus"] = ifOpStatus
	new_row["ifOpStatus__descr"] = interfaceStatusString(ifOpStatus)
	new_row["ifStatus"] = calcStatus(ifAdminStatus, ifOpStatus)
	new_row["ifStatus__descr"] = statusString(ifAdminStatus, ifOpStatus)

	new_row["ifOpStatus"] = GetInt32WithDefault(params, res, "1.3.6.1.2.1.2.2.1.8."+ifIndex, -1)
	new_row["ifLastChange"] = GetUint32WithDefault(params, res, "1.3.6.1.2.1.2.2.1.9."+ifIndex, 0)
	new_row["ifInOctets"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.10."+ifIndex, 0)
	new_row["ifInUcastPkts"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.11."+ifIndex, 0)
	new_row["ifInNUcastPkts"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.12."+ifIndex, 0)
	new_row["ifInDiscards"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.13."+ifIndex, 0)
	new_row["ifInErrors"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.14."+ifIndex, 0)
	new_row["ifInUnknownProtos"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.15."+ifIndex, 0)
	new_row["ifOutOctets"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.16."+ifIndex, 0)
	new_row["ifOutUcastPkts"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.17."+ifIndex, 0)
	new_row["ifOutNUcastPkts"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.18."+ifIndex, 0)
	new_row["ifOutDiscards"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.19."+ifIndex, 0)
	new_row["ifOutErrors"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.20."+ifIndex, 0)
	new_row["ifOutQLen"] = GetUint64WithDefault(params, res, "1.3.6.1.2.1.2.2.1.21."+ifIndex, 0)
	new_row["ifSpecific"] = GetOidWithDefault(params, res, "1.3.6.1.2.1.2.2.1.22."+ifIndex)

	return new_row, nil
}

type portStatus struct {
	snmpBase
}

func (self *portStatus) Call(params MContext) (interface{}, error) {
	ifIndex, e := params.GetString("@ifIndex")
	if nil != e || 0 == len(ifIndex) {
		return nil, IsRequired("ifIndex")
	}

	oidBuffer := bytes.NewBuffer(make([]byte, 0, 200))
	for _, s := range []string{
		"1.3.6.1.2.1.2.2.1.7.",
		"1.3.6.1.2.1.2.2.1.8."} {
		oidBuffer.WriteString(s)
		oidBuffer.WriteString(ifIndex)
		oidBuffer.WriteString(",")
	}
	oidBuffer.Truncate(oidBuffer.Len() - 1)

	res, e := self.Get(params, oidBuffer.String())
	if nil != e {
		return nil, e
	}

	new_row := map[string]interface{}{}
	new_row["ifIndex"] = ifIndex
	ifAdminStatus := GetInt32WithDefault(params, res, "7", -1)
	ifOpStatus := GetInt32WithDefault(params, res, "8", -1)
	new_row["ifAdminStatus"] = ifAdminStatus
	new_row["ifAdminStatus__descr"] = interfaceStatusString(ifAdminStatus)
	new_row["ifOpStatus"] = ifOpStatus
	new_row["ifOpStatus__descr"] = interfaceStatusString(ifOpStatus)
	new_row["ifStatus"] = calcStatus(ifAdminStatus, ifOpStatus)
	new_row["ifStatus__descr"] = statusString(ifAdminStatus, ifOpStatus)

	return new_row, nil
}

type portScalar struct {
	snmpBase
}

func (self *portScalar) Call(params MContext) (interface{}, error) {
	ifIndex, e := params.GetString("@ifIndex")
	if nil != e || 0 == len(ifIndex) {
		return nil, IsRequired("ifIndex")
	}
	return self.CallWithIfIndex(params, ifIndex)
}

func (self *portScalar) CallWithIfIndex(params MContext, ifIndex string) (interface{}, error) {

	oidBuffer := bytes.NewBuffer(make([]byte, 0, 900))
	for _, s := range []string{"1.3.6.1.2.1.2.2.1.7.",
		"1.3.6.1.2.1.2.2.1.8.",
		"1.3.6.1.2.1.2.2.1.10.",
		"1.3.6.1.2.1.2.2.1.11.",
		"1.3.6.1.2.1.2.2.1.12.",
		"1.3.6.1.2.1.2.2.1.13.",
		"1.3.6.1.2.1.2.2.1.14.",
		"1.3.6.1.2.1.2.2.1.15.",
		"1.3.6.1.2.1.2.2.1.16.",
		"1.3.6.1.2.1.2.2.1.17.",
		"1.3.6.1.2.1.2.2.1.18.",
		"1.3.6.1.2.1.2.2.1.19.",
		"1.3.6.1.2.1.2.2.1.20."} {
		oidBuffer.WriteString(s)
		oidBuffer.WriteString(ifIndex)
		oidBuffer.WriteString(",")
	}
	oidBuffer.Truncate(oidBuffer.Len() - 1)

	old_row, e := self.Get(params, oidBuffer.String())
	if nil != e {
		return nil, e
	}

	new_row := map[string]interface{}{}
	new_row["ifIndex"] = ifIndex
	new_row["ifBit"] = 32

	ifAdminStatus := GetInt32WithDefault(params, old_row, "7", -1)
	ifOpStatus := GetInt32WithDefault(params, old_row, "8", -1)
	new_row["ifAdminStatus"] = ifAdminStatus
	new_row["ifAdminStatus__descr"] = interfaceStatusString(ifAdminStatus)
	new_row["ifOpStatus"] = ifOpStatus
	new_row["ifOpStatus__descr"] = interfaceStatusString(ifOpStatus)
	new_row["ifStatus"] = calcStatus(ifAdminStatus, ifOpStatus)
	new_row["ifStatus__descr"] = statusString(ifAdminStatus, ifOpStatus)

	new_row["ifInOctets"] = GetUint64WithDefault(params, old_row, "10", 0)
	new_row["ifInUcastPkts"] = GetUint64WithDefault(params, old_row, "11", 0)
	new_row["ifInNUcastPkts"] = GetUint64WithDefault(params, old_row, "12", 0)
	new_row["ifInDiscards"] = GetUint64WithDefault(params, old_row, "13", 0)
	new_row["ifInErrors"] = GetUint64WithDefault(params, old_row, "14", 0)
	new_row["ifInUnknownProtos"] = GetUint64WithDefault(params, old_row, "15", 0)
	new_row["ifOutOctets"] = GetUint64WithDefault(params, old_row, "16", 0)
	new_row["ifOutUcastPkts"] = GetUint64WithDefault(params, old_row, "17", 0)
	new_row["ifOutNUcastPkts"] = GetUint64WithDefault(params, old_row, "18", 0)
	new_row["ifOutDiscards"] = GetUint64WithDefault(params, old_row, "19", 0)
	new_row["ifOutErrors"] = GetUint64WithDefault(params, old_row, "20", 0)
	return new_row, nil
}

type portDescr struct {
	snmpBase
}

func (self *portDescr) Call(params MContext) (interface{}, error) {

	ifIndex, e := params.GetString("@ifIndex")
	if nil != e || 0 == len(ifIndex) {
		return nil, IsRequired("ifIndex")
	}

	oidBuffer := bytes.NewBuffer(make([]byte, 0, 900))
	for _, s := range []string{"1.3.6.1.2.1.2.2.1.2.",
		"1.3.6.1.2.1.2.2.1.3.",
		"1.3.6.1.2.1.2.2.1.4.",
		"1.3.6.1.2.1.2.2.1.5.",
		"1.3.6.1.2.1.2.2.1.6.",
		"1.3.6.1.2.1.2.2.1.22."} {
		oidBuffer.WriteString(s)
		oidBuffer.WriteString(ifIndex)
		oidBuffer.WriteString(",")
	}
	oidBuffer.Truncate(oidBuffer.Len() - 1)

	old_row, e := self.Get(params, oidBuffer.String())
	if nil != e {
		return nil, e
	}

	new_row := map[string]interface{}{}
	new_row["ifIndex"] = ifIndex
	new_row["ifDescr"] = GetStringWithDefault(params, old_row, "2")
	new_row["ifType"] = GetInt32WithDefault(params, old_row, "3", -1)
	new_row["ifMtu"] = GetInt32WithDefault(params, old_row, "4", -1)
	new_row["ifSpeed"] = GetUint64WithDefault(params, old_row, "5", 0)
	new_row["ifPhysAddress"] = GetHardwareAddressWithDefault(params, old_row, "6")
	new_row["ifSpecific"] = GetOidWithDefault(params, old_row, "22")
	return new_row, nil
}

type interfaceAll struct {
	snmpBase
}

func (self *interfaceAll) Call(params MContext) (interface{}, error) {
	return self.GetAllResult(params, "1.3.6.1.2.1.2.2.1", "1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := readInterface(params, old_row)
			new_row["ifIndex"] = GetInt32WithDefault(params, old_row, "1", -1)
			return new_row, nil
		})
}

type interfaceStatus struct {
	snmpBase
}

func (self *interfaceStatus) Call(params MContext) (interface{}, error) {
	return self.GetAllResult(params, "1.3.6.1.2.1.2.2.1", "1,7,8",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["ifIndex"] = GetInt32WithDefault(params, old_row, "1", -1)

			ifAdminStatus := GetInt32WithDefault(params, old_row, "7", -1)
			ifOpStatus := GetInt32WithDefault(params, old_row, "8", -1)
			new_row["ifAdminStatus"] = ifAdminStatus
			new_row["ifAdminStatus__descr"] = interfaceStatusString(ifAdminStatus)
			new_row["ifOpStatus"] = ifOpStatus
			new_row["ifOpStatus__descr"] = interfaceStatusString(ifOpStatus)
			new_row["ifStatus"] = calcStatus(ifAdminStatus, ifOpStatus)
			new_row["ifStatus__descr"] = statusString(ifAdminStatus, ifOpStatus)

			return new_row, nil
		})
}

type interfaceScalar struct {
	snmpBase
}

func (self *interfaceScalar) Call(params MContext) (interface{}, error) {
	return self.GetAllResult(params, "1.3.6.1.2.1.2.2.1", "1,7,8,10,11,12,13,14,15,16,17,18,19,20",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["ifBit"] = 32
			new_row["ifIndex"] = GetInt32WithDefault(params, old_row, "1", -1)

			ifAdminStatus := GetInt32WithDefault(params, old_row, "7", -1)
			ifOpStatus := GetInt32WithDefault(params, old_row, "8", -1)
			new_row["ifAdminStatus"] = ifAdminStatus
			new_row["ifAdminStatus__descr"] = interfaceStatusString(ifAdminStatus)
			new_row["ifOpStatus"] = ifOpStatus
			new_row["ifOpStatus__descr"] = interfaceStatusString(ifOpStatus)
			new_row["ifStatus"] = calcStatus(ifAdminStatus, ifOpStatus)
			new_row["ifStatus__descr"] = statusString(ifAdminStatus, ifOpStatus)

			new_row["ifInOctets"] = GetUint64WithDefault(params, old_row, "10", 0)
			new_row["ifInUcastPkts"] = GetUint64WithDefault(params, old_row, "11", 0)
			new_row["ifInNUcastPkts"] = GetUint64WithDefault(params, old_row, "12", 0)
			new_row["ifInDiscards"] = GetUint64WithDefault(params, old_row, "13", 0)
			new_row["ifInErrors"] = GetUint64WithDefault(params, old_row, "14", 0)
			new_row["ifInUnknownProtos"] = GetUint64WithDefault(params, old_row, "15", 0)
			new_row["ifOutOctets"] = GetUint64WithDefault(params, old_row, "16", 0)
			new_row["ifOutUcastPkts"] = GetUint64WithDefault(params, old_row, "17", 0)
			new_row["ifOutNUcastPkts"] = GetUint64WithDefault(params, old_row, "18", 0)
			new_row["ifOutDiscards"] = GetUint64WithDefault(params, old_row, "19", 0)
			new_row["ifOutErrors"] = GetUint64WithDefault(params, old_row, "20", 0)
			return new_row, nil
		})
}

type interfaceDescr struct {
	snmpBase
}

func (self *interfaceDescr) Call(params MContext) (interface{}, error) {
	return self.GetAllResult(params, "1.3.6.1.2.1.2.2.1", "1,2,3,4,5,6,22",
		func(key string, old_row map[string]interface{}) (map[string]interface{}, error) {
			new_row := map[string]interface{}{}
			new_row["ifIndex"] = GetInt32WithDefault(params, old_row, "1", -1)
			new_row["ifDescr"] = GetStringWithDefault(params, old_row, "2")
			ifType := GetInt32WithDefault(params, old_row, "3", -1)
			new_row["ifType"] = ifType
			new_row["ifType__descr"] = interfaceType(ifType)
			new_row["ifMtu"] = GetInt32WithDefault(params, old_row, "4", -1)
			new_row["ifSpeed"] = GetUint64WithDefault(params, old_row, "5", 0)
			new_row["ifPhysAddress"] = GetHardwareAddressWithDefault(params, old_row, "6")
			new_row["ifSpecific"] = GetOidWithDefault(params, old_row, "22")
			return new_row, nil
		})
}

func init() {
	Methods["port_interface"] = newRouteSpecWithPaths("get", "interface", "the interface info", []P{P{"port", "@ifIndex"}}, nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &portAll{}
			return drv, drv.Init(params)
		})

	Methods["port_interface_descr"] = newRouteSpecWithPaths("get", "interface_descr", "the descr part of interface info", []P{P{"port", "@ifIndex"}}, nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &portDescr{}
			return drv, drv.Init(params)
		})

	Methods["port_interface_status"] = newRouteSpecWithPaths("get", "interface_status", "the status part of interface info", []P{P{"port", "@ifIndex"}}, nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &portStatus{}
			return drv, drv.Init(params)
		})

	Methods["port_interface_scalar"] = newRouteSpecWithPaths("get", "interface_scalar", "the scalar part of interface info", []P{P{"port", "@ifIndex"}}, nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &portScalar{}
			return drv, drv.Init(params)
		})

	Methods["default_interface"] = newRouteSpec("get", "interface", "the interface info", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &interfaceAll{}
			return drv, drv.Init(params)
		})

	Methods["default_interface_descr"] = newRouteSpec("get", "interface_descr", "the descr part of interface info", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &interfaceDescr{}
			return drv, drv.Init(params)
		})

	Methods["default_interface_status"] = newRouteSpec("get", "interface_status", "the status part of interface info", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &interfaceStatus{}
			return drv, drv.Init(params)
		})

	Methods["default_interface_scalar"] = newRouteSpec("get", "interface_scalar", "the scalar part of interface info", nil,
		func(rs *RouteSpec, params map[string]interface{}) (Method, error) {
			drv := &interfaceScalar{}
			return drv, drv.Init(params)
		})
}
