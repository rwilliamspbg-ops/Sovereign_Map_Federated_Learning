import json,sys,time,urllib.request,os,subprocess
out_dir=sys.argv[1]
os.makedirs(out_dir,exist_ok=True)
progress=os.path.join(out_dir,'round_progress.csv')
with open(progress,'w') as f:
    f.write('timestamp,round,accuracy,loss\n')
start=time.time(); max_wait=60*120

def get_json(url):
    with urllib.request.urlopen(url, timeout=8) as r:
        return json.loads(r.read().decode())

reached=0
while True:
    now=int(time.time())
    try:
        c=get_json('http://localhost:8000/convergence')
        r=int(c.get('current_round',0)); a=c.get('current_accuracy',0); l=c.get('current_loss',0)
    except Exception:
        r=0; a=0; l=0
    with open(progress,'a') as f:
        f.write(f"{now},{r},{a},{l}\n")
    print(f"sample round={r}", flush=True)
    if r>=200:
        reached=1
        break
    if time.time()-start>=max_wait:
        break
    time.sleep(15)

for name,url in [
    ('convergence_final.json','http://localhost:8000/convergence'),
    ('metrics_summary_final.json','http://localhost:8000/metrics_summary'),
    ('prometheus_metrics_snapshot.txt','http://localhost:8000/metrics'),
    ('prometheus_targets.json','http://localhost:9090/api/v1/targets'),
    ('prometheus_rules.json','http://localhost:9090/api/v1/rules'),
]:
    try:
        with urllib.request.urlopen(url, timeout=20) as resp:
            body=resp.read().decode()
        with open(os.path.join(out_dir,name),'w') as f:
            f.write(body)
    except Exception:
        pass

for svc,tail in [('backend','4000'),('node-agent','2500'),('prometheus','1200'),('grafana','1200'),('alertmanager','1200')]:
    try:
        with open(os.path.join(out_dir,f'{svc}.log'),'w') as f:
            f.write(subprocess.check_output(['docker','compose','-f','docker-compose.full.yml','logs','--tail='+tail,svc], text=True, cwd='/workspaces/Sovereign_Map_Federated_Learning'))
    except Exception:
        pass

try:
    with open(os.path.join(out_dir,'compose-ps.txt'),'w') as f:
        f.write(subprocess.check_output(['docker','compose','-f','docker-compose.full.yml','ps'], text=True, cwd='/workspaces/Sovereign_Map_Federated_Learning'))
except Exception:
    pass

last_round=0; last_acc=0; last_loss=0; samples=0
with open(progress) as f:
    next(f,None)
    for line in f:
        samples += 1
        _,r,a,l=line.strip().split(',',3)
        last_round=int(float(r)); last_acc=a; last_loss=l
with open(os.path.join(out_dir,'summary.txt'),'w') as f:
    f.write(f"target_round_reached={reached}\n")
    f.write(f"samples={samples}\n")
    f.write(f"final_round={last_round}\n")
    f.write(f"final_accuracy={last_acc}\n")
    f.write(f"final_loss={last_loss}\n")
print(f"target_round_reached={reached}")
print(f"final_round={last_round}")
print(f"artifacts_ready={out_dir}")
