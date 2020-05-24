.PHONY: all clean

APP_DIR			:= .
APP_INC_DIR		:= $(APP_DIR)/include
APP_LIB_DIR		:= $(APP_DIR)/lib

livestream:
	CGO_ENABLED=1  CGO_LDFLAGS="-L$(APP_LIB_DIR)" go build -i -o livestream -gcflags "-N"

clean:
	rm -f livestream
	rm -f /home/user/huawei/share/livestrem

	
	





