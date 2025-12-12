import requests
import jwt
import uuid

def create_jwt():
    jwt_key = ""
    jf = open("key.txt", "r")
    jwt_key = jf.read()
    jf.close()
    claims = {
        "email": "vivienne@kthais.com",
        "roles": ["admin"],
    }
    token = jwt.encode(claims, jwt_key, algorithm="HS256")
    return token


def upload_companies(file_path, api_url):
    token = create_jwt() 
    with open(file_path, 'rb') as file:
        headers = {"authorization": f"Bearer {token}"}
        for line in file.readlines():
            name, description, logo = line.decode('utf-8').strip().split(',')
            id = uuid.uuid4()
            logo_id = uuid.uuid4()
            data = {
                "id": str(id),
                "name": name,
                "description": description,
                "logo": str(logo_id)
            }
            response = requests.post(f"{api_url}/addCompany", json=data, headers=headers)
            if response.ok:
                print(f"Uploaded company: {name}")
            else:
                print(f"Failed to upload company: {name}, Status Code: {response.status_code}, Response: {response.text}")


def get_companies(api_url, save=False):
    token = create_jwt()
    headers = {"authorization": token}
    resp = requests.get(f"{api_url}/getAllCompanies", headers=headers)
    if resp.ok:
        names = []
        with open("companies.csv", "rb") as f:
            for line in f.readlines():
                name, description, logo = line.decode('utf-8').strip().split(',')
                names.append(name)
        companies = resp.json()
        c_names = [c['name'] for c in companies]
        success = True
        for name in names:
            if name not in c_names:
                print(f"Company {name} not found in API response")
                success = False
        if success:
            print("All companies verified successfully.")
        if save:
            with open("output_companies.json", "w") as out_file:
                import json
                json.dump(companies, out_file, indent=4)
    else:
        print(f"Failed to get companies, Status Code: {resp.status_code}, Response: {resp.text}")


def get_specific(api_url):
    token = create_jwt()
    headers = {"authorization": token}
    resp = requests.get(f"{api_url}/getAllCompanies", headers=headers)
    if resp.ok:
        names = []
        with open("companies.csv", "rb") as f:
            for line in f.readlines():
                name, description, logo = line.decode('utf-8').strip().split(',')
                names.append(name)
        companies = resp.json()
        id = companies[0]['id']
        params = {"id": id}
        resp2 = requests.get(f"{api_url}/getCompany", headers=headers, params=params)
        if resp2.ok:
            company = resp2.json()
            print(f"Retrieved company: {company}")


if __name__ == "__main__":
    api_url = "http://localhost:8080/api/v1/company"
    filepath = "./companies.csv"
    # upload_companies(filepath, api_url)
    get_companies(api_url, save=True)
    get_specific(api_url)