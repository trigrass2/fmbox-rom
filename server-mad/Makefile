
all:
	arm-linux-gcc -static -o mad-arm mad.c -lmad 
	gcc -static -o mad-i386 mad.c -lmad 

test:
	./mad-i386 < /var/www/test.mp3 >/tmp/out 2>log

clean:
	rm mad-arm mad-i386

