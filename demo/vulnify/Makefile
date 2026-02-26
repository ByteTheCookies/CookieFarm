build: vulnify.c
	gcc -o vulnify -pie -fstack-protector vulnify.c && strip -s vulnify

debug: vulnify.c
	gcc -o vulnify -Wno-format-truncation -Wall -Wextra -Werror vulnify.c

