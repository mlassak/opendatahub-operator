import yaml
import os
import sys

def verify_pr_author():
    pr_author = os.environ.get('PR_AUTHOR')

    try:
        with open('OWNERS_ALIASES', 'r') as f:
            owners_data = yaml.safe_load(f)
        platform_owners = owners_data.get('aliases', {}).get('platform', [])
        if pr_author in platform_owners:
            print(f"Author {pr_author} found in OWNERS file. Proceeding.")
            sys.exit(0)
        else:
            print(f"Author {pr_author} is not in the OWNERS_ALIASES file.")
            sys.exit(1)
    except FileNotFoundError:
        print("OWNERS_ALIASES file not found.")
        sys.exit(1)
    except Exception as e:
        print(f"Error occurred while parsing OWNERS_ALIASES file: {e}")
        sys.exit(1)

if __name__ == "__main__":
    verify_pr_author()
