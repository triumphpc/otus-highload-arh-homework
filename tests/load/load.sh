for i in {1..1000}; do
  first_name=$(cat /dev/urandom | tr -dc 'a-zA-Z' | fold -w 5 | head -n 1)
  last_name=$(cat /dev/urandom | tr -dc 'a-zA-Z' | fold -w 8 | head -n 1)

  hey -n 1000 -c 10 \
    -H "Accept: application/json" \
    -H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDgyMzgzMjQsImlhdCI6MTc0ODIzNjUyNCwic3ViIjoyMDAwNH0.kHRnZZ34_fcqUrLUb2vK3Rw2JJbJv7rOWZFG8O6wQow" \
    "http://localhost:8080/api/v1/user/search?first_name=$first_name&last_name=$last_name" &
done

wait