# MP0

```bash
# Client Test
./bin/mp0-s 8080 or socat TCP-LISTEN:8080 -
./bin/mp0-c A 127.0.0.1 8080
python3 ./script/mp0/generator.py 1 100 | ./bin/mp0-c A 127.0.0.1 8080
```
