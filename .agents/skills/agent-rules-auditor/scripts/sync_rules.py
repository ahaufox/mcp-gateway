import os
import glob

MASTER_DIR = ".agents/rules"
TARGETS = {
    "Cursor": (".cursor/rules", ".mdc"),
    "Windsurf": (".windsurf/rules", ".md"),
    "Claude": (".claude/rules", ".md"),
    "Trae": (".trae/rules", ".md"),
}

def strip_fm(content):
    if content.startswith('---'):
        parts = content.split('---', 2)
        if len(parts) >= 3:
            return parts[0] + '---' + parts[1] + '---', parts[2].strip()
    return "", content.strip()

def sync():
    master_files = glob.glob(os.path.join(MASTER_DIR, "*.md"))
    master_files = [f for f in master_files if os.path.basename(f) != "README.md"]

    for master_path in master_files:
        filename = os.path.basename(master_path)
        with open(master_path, 'r', encoding='utf-8') as f:
            m_content = f.read()
        
        m_fm, m_body = strip_fm(m_content)

        for tool, (dir_path, ext) in TARGETS.items():
            if not os.path.exists(dir_path):
                continue
            
            target_filename = filename.replace(".md", ext)
            target_path = os.path.join(dir_path, target_filename)

            if not os.path.exists(target_path):
                print(f"Creating missing {tool} rule: {target_filename}")
                # For missing, use master FM if it exists, but Cursor might need more.
                # Just use master content as is.
                with open(target_path, 'w', encoding='utf-8') as f:
                    f.write(m_content)
                continue

            with open(target_path, 'r', encoding='utf-8') as f:
                t_content = f.read()
            
            t_fm, t_body = strip_fm(t_content)

            if m_body != t_body:
                print(f"Updating {tool} rule: {target_filename}")
                # Preserve target FM, use master body
                new_content = t_fm + "\n\n" + m_body + "\n"
                with open(target_path, 'w', encoding='utf-8') as f:
                    f.write(new_content)

if __name__ == "__main__":
    sync()
