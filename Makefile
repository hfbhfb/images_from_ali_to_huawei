
run:
	@echo "编译和运行镜像迁移工具:阿里云->华为云"
	go build -o images_from_ali_to_huawei main.go
	./images_from_ali_to_huawei --config=./config.json

clean:
	@echo "清空包含有敏感信息的config.json文件"
	rm config.json

