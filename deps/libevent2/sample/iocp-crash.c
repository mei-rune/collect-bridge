
#include <string.h>
#include <errno.h>
#include <stdio.h>
#include <signal.h>

#include <event2/listener.h>
#include <event2/util.h>
#include <event2/event.h>


static const int PORT = 9995;


static void
listener_cb(struct evconnlistener *listener, evutil_socket_t fd,
    struct sockaddr *sa, int socklen, void *user_data)
{
}


int
main(int argc, char **argv)
{
	struct event_config *evcfg;
	struct event_base *base;
	struct evconnlistener *listener;

	struct sockaddr_in sin;
	WSADATA wsa_data;
	WSAStartup(0x0201, &wsa_data);
		
	evcfg = event_config_new();
	event_config_set_flag(evcfg, EVENT_BASE_FLAG_STARTUP_IOCP);
	base = event_base_new_with_config(evcfg);
	//base = event_base_new();
	if (!base) {
		fprintf(stderr, "Could not initialize libevent!\n");
		return 1;
	}

	memset(&sin, 0, sizeof(sin));
	sin.sin_family = AF_INET;
	sin.sin_port = htons(PORT);

	listener = evconnlistener_new_bind(base, listener_cb, (void *)base,
	    LEV_OPT_REUSEABLE|LEV_OPT_CLOSE_ON_FREE, -1,
	    (struct sockaddr*)&sin,
	    sizeof(sin));

	if (!listener) {
		fprintf(stderr, "Could not create a listener!\n");
		return 1;
	}


	event_base_dispatch(base);
	return 0;
}
