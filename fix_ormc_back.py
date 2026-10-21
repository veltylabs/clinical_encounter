with open("model.go", "r") as f:
    text = f.read()

text = text.replace('{Name: "status", Type: input.Select(), NotNull: true},', '{Name: "status", Type: statusSelectInput(), NotNull: true},')

with open("model.go", "w") as f:
    f.write(text)
