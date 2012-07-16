{
  'targets': [
    {
      'target_name': 'collect-bridge',
      'type': 'executable',
      'dependencies': [
        'deps/libevent2.gyp:libevent2',
        'deps/libevent2.gyp:libevent2_core',
        'deps/libevent2.gyp:libevent2_extras',
        'deps/libssh2.gyp:libssh2',
        'deps/libMU/libMU.gyp:libMU',
      ],
      'sources': [
      ],
      'msvs-settings': {
        'VCLinkerTool': {
          'SubSystem': 1, # /subsystem:console
        },
      },
      'conditions': [
        ['OS == "linux"', {
          'libraries': ['-ldl'],
        }],
        ['OS=="linux" or OS=="freebsd" or OS=="openbsd" or OS=="solaris"', {
          'cflags': [ '--std=c89' ],
          'defines': [ '_GNU_SOURCE' ]
        }],
      ],
      'defines': [ 'BUNDLE=1' ]
    },
   
  ],
}
