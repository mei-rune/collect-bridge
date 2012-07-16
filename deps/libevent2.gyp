{
  'variables': {
    'YAJL_MAJOR': '2',
    'YAJL_MINOR': '0',
    'YAJL_MICRO': '5',
  },
  'targets': [
    {
      'target_name': 'libevent2',
      'type': 'static_library',
      'sources': [
          'libevent2/event.c',
          'libevent2/buffer.c',
          'libevent2/bufferevent.c',
          'libevent2/bufferevent_sock.c',
          'libevent2/bufferevent_pair.c',
          'libevent2/listener.c',
          'libevent2/evmap.c',
          'libevent2/log.c',
          'libevent2/evutil.c',
          'libevent2/strlcpy.c',
          'libevent2/signal.c',
          'libevent2/bufferevent_filter.c',
          'libevent2/evthread.c',
          'libevent2/bufferevent_ratelim.c',
          'libevent2/evutil_rand.c',
          'libevent2/event_tagging.c',
          'libevent2/http.c',
          'libevent2/evdns.c',
          'libevent2/evrpc.c',
      ],
      'direct_dependent_settings': {
        'include_dirs': [
          '<(SHARED_INTERMEDIATE_DIR)',
        ]
      },
      'include_dirs': [
        'libevent2/include',
        '<(SHARED_INTERMEDIATE_DIR)',
      ],
      'defines': [ 'NETSNMP_ENABLE_IPV6' ],
       'conditions': [
			['OS=="linux" or OS=="freebsd" or OS=="openbsd" or OS=="solaris"', {
			  'cflags': [ '--std=c89' ],
			  'defines': [ '_GNU_SOURCE' ]
			}],
			['OS=="win"', {
			  'defines': [ 'HAVE_WIN32_PLATFORM_SDK' ],
			  'sources': [
				  'libevent2/win32select.c',
				  'libevent2/evthread_win32.c',
				  'libevent2/buffer_iocp.c',
				  'libevent2/event_iocp.c',
				  'libevent2/bufferevent_async.c',
			  ],      
			  'include_dirs': [
				'libevent2/compat',
				'libevent2/WIN32-Code',
			  ],
			}],
       ],
    }, # end libevent2
    {
      'target_name': 'libevent2_core',
      'type': 'static_library',
      'sources': [
          'libevent2/event.c',
          'libevent2/buffer.c',
          'libevent2/bufferevent.c',
          'libevent2/bufferevent_sock.c',
          'libevent2/bufferevent_pair.c',
          'libevent2/listener.c',
          'libevent2/evmap.c',
          'libevent2/log.c',
          'libevent2/evutil.c',
          'libevent2/strlcpy.c',
          'libevent2/signal.c',
          'libevent2/bufferevent_filter.c',
          'libevent2/evthread.c',
          'libevent2/bufferevent_ratelim.c',
          'libevent2/evutil_rand.c',
      ],
      'direct_dependent_settings': {
        'include_dirs': [
          '<(SHARED_INTERMEDIATE_DIR)',
        ]
      },
      'include_dirs': [
        'libevent2/include',
        '<(SHARED_INTERMEDIATE_DIR)',
      ],
      'defines': [ 'NETSNMP_ENABLE_IPV6' ],
       'conditions': [
			['OS=="linux" or OS=="freebsd" or OS=="openbsd" or OS=="solaris"', {
			  'cflags': [ '--std=c89' ],
			  'defines': [ '_GNU_SOURCE' ]
			}],
			['OS=="win"', {
			  'defines': [ 'HAVE_WIN32_PLATFORM_SDK' ],
			  'sources': [
				  'libevent2/win32select.c',
				  'libevent2/evthread_win32.c',
				  'libevent2/buffer_iocp.c',
				  'libevent2/event_iocp.c',
				  'libevent2/bufferevent_async.c',
			  ],      
			  'include_dirs': [
				'libevent2/compat',
				'libevent2/WIN32-Code',
			  ],
			}],
       ],
    }, # end libevent2_core
    {
      'target_name': 'libevent2_extras',
      'type': 'static_library',
      'sources': [
          'libevent2/event_tagging.c',
          'libevent2/http.c',
          'libevent2/evdns.c',
          'libevent2/evrpc.c',
      ],
      'direct_dependent_settings': {
        'include_dirs': [
          '<(SHARED_INTERMEDIATE_DIR)',
        ]
      },
      'include_dirs': [
        'libevent2/include',
        '<(SHARED_INTERMEDIATE_DIR)',
      ],
      'defines': [ 'NETSNMP_ENABLE_IPV6' ],
       'conditions': [
			['OS=="linux" or OS=="freebsd" or OS=="openbsd" or OS=="solaris"', {
			  'cflags': [ '--std=c89' ],
			  'defines': [ '_GNU_SOURCE' ]
			}],
			['OS=="win"', {
			  'defines': [ 'HAVE_WIN32_PLATFORM_SDK' ],    
			  'include_dirs': [
				'libevent2/compat',
				'libevent2/WIN32-Code',
			  ],
			}],
       ],
    }, # end libevent2_extras
  ] # end targets
}
