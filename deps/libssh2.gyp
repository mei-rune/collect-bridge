{
  'variables': {
    'YAJL_MAJOR': '2',
    'YAJL_MINOR': '0',
    'YAJL_MICRO': '5',
  },
  'targets': [
    {
      'target_name': 'libssh2',
      'type': 'static_library',
      'sources': [
          'libssh2/src/agent.c',
          'libssh2/src/channel.c',
          'libssh2/src/comp.c',
          'libssh2/src/crypt.c',
          'libssh2/src/global.c',
          'libssh2/src/hostkey.c',
          'libssh2/src/keepalive.c',
          'libssh2/src/kex.c',
          'libssh2/src/knownhost.c',
          'libssh2/src/libgcrypt.c',
          'libssh2/src/mac.c',
          'libssh2/src/misc.c',
          'libssh2/src/openssl.c',
          'libssh2/src/packet.c',
          'libssh2/src/pem.c',
          'libssh2/src/publickey.c',
          'libssh2/src/scp.c',
          'libssh2/src/session.c',
          'libssh2/src/sftp.c',
          'libssh2/src/transport.c',
          'libssh2/src/userauth.c',
          'libssh2/src/version.c',
      ],
      'dependencies': [
		'openssl/openssl.gyp:openssl'
	  ],
	  'export_dependent_settings': [
		'openssl/openssl.gyp:openssl'
	  ],
      'direct_dependent_settings': {
        'include_dirs': [
          '<(SHARED_INTERMEDIATE_DIR)',
        ]
      },
      'include_dirs': [
        'libssh2/src',
        'libssh2/win32',
        'libssh2/include',
        '<(SHARED_INTERMEDIATE_DIR)',
      ],
      'defines': [ 'NETSNMP_ENABLE_IPV6' ],
      'conditions': [
			['OS=="linux" or OS=="freebsd" or OS=="openbsd" or OS=="solaris"', {
			  'cflags': [ '--std=c89' ],
			  'defines': [ '_GNU_SOURCE' ]
			}],
			['OS=="win32"', {
			  'defines': [ 'HAVE_WIN32_PLATFORM_SDK' ]
			}],
       ],
    }, # end libssh2
  ] # end targets
}
