//run in nodejs
import OSS from 'ali-oss';
import { readFileSync } from 'fs';
import { resolve } from 'path';
import { request } from 'https';
import { spawn } from 'child_process';
import glob from 'tiny-glob';
import { web as webDeploymentConfig } from '../../.deployment.cjs'

const distDir = resolve(process.cwd(), './dist')

const rsyncDest = webDeploymentConfig.rsync?.destination
if (rsyncDest) {
  const sync1 = spawn(`rsync -avz --delete --exclude "*.map" ${distDir}/. ${rsyncDest}`, {
    shell: true,
    stdio: 'pipe',
  })
  sync1.stdin.end()
  sync1.stdout.pipe(process.stdout)
  sync1.stderr.pipe(process.stderr)
}

uploadToOSS();
async function uploadToOSS() {
  const publicPath = webDeploymentConfig.publicPath;
  const { cdnDest, ossConfig } = webDeploymentConfig.cdn || {}
  if (!cdnDest || !ossConfig || !publicPath) return;

  const store = new OSS(ossConfig);

  let left = 0;
  const files = await glob(`${distDir}/**/*`, { absolute: true, filesOnly: true });
  for (const filePath of files) {
    if (/\.(map|html|LICENSE\.txt)$/.test(filePath)) continue;
    const basePath = filePath.replace(distDir, '').slice(1);
    const ossPath = `${cdnDest}/${basePath}`;
    const url = `${publicPath}${basePath}`;

    left++;
    store.put(ossPath, readFileSync(filePath)).then(() => {
      console.log(`uploaded ${basePath}`);
      const req = request(new URL(url), function (res) {
        console.log(`validate [ ${url} ]`, res.statusCode);
        setTimeout(() => res.destroy(), 100);
        left--;
        if (left === 0) console.log('all done');
      })
      req.end()
    });
  }
}
