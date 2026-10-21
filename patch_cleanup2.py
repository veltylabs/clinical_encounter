import re

with open("model.go", "r") as f:
    text = f.read()

# Let's ensure CreateVisitArgsModel actually HAS service_id, not double or none
# In the previous python replacement I may have stripped it completely or not added it.
if '{Name: "service_id", Type: input.Text()},' not in text:
    # Need to put it back properly
    pass

# Wait, let's just make sure CreateVisitArgsModel has it exactly once.
lines = text.split("\n")
out = []
in_create = False
has_res = False
for line in lines:
    if 'Name: "create_visit_args",' in line:
        in_create = True
    elif 'func statusSelectInput' in line:
        in_create = False

    if in_create and '{Name: "reservation_id", Type: input.Text()},' in line:
        # Check if next line has service_id
        out.append(line)
        out.append('\t\t{Name: "service_id", Type: input.Text()},')
    elif in_create and '{Name: "service_id", Type: input.Text()},' in line:
        # omit it to avoid duplicates
        pass
    else:
        out.append(line)

with open("model.go", "w") as f:
    f.write("\n".join(out))
