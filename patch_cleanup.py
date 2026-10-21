import re

# 1. Clean up duplicate definitions in supply.go, ops.go
with open("supply.go", "r") as f:
    text = f.read()

# remove structs that ormc generated
text = re.sub(r'type UpsertSupplyUseArgs struct \{[\s\S]*?\}', '', text)
text = re.sub(r'type SupplyUseRefArgs struct \{[\s\S]*?\}', '', text)
text = re.sub(r'type ListSupplyUseArgs struct \{[\s\S]*?\}', '', text)
text = re.sub(r'type ListPendingSupplyArgs struct \{[\s\S]*?\}', '', text)
with open("supply.go", "w") as f:
    f.write(text)

with open("ops.go", "r") as f:
    text = f.read()
text = re.sub(r'type CompleteVisitArgs struct \{[\s\S]*?\}', '', text)
with open("ops.go", "w") as f:
    f.write(text)

# 2. Fix the duplicate service_id in model.go which causes duplicate ServiceId in model_orm.go
with open("model.go", "r") as f:
    text = f.read()

text = text.replace('{Name: "service_id", Type: input.Text()},\n\t\t{Name: "service_id", Type: input.Text()},', '{Name: "service_id", Type: input.Text()},')
with open("model.go", "w") as f:
    f.write(text)
