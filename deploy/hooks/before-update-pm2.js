const fs = require("fs");
const { SERVICE_NAME, APP_DIR } = process.env;

var pm2configPath = `${APP_DIR}/ecosystem.config.js`;
var content = fs.readFileSync(pm2configPath, "utf8");

content = content.replace(/SERVICE_NAME/g, SERVICE_NAME);
content = content.replace(/\/\*+\s*@import:\s*(\S+)\s*\*\//g, (_, envFilePath) => {
  var envFileContent = fs.readFileSync(envFilePath, "utf8");
  var output = [_];
  envFileContent.split("\n").forEach(line => {
    line = line.trim();
    if (!line.startsWith("#") && line.includes("=")) {
      var idx = line.indexOf("=");
      var key = line.substring(0, idx).trim();
      var value = line.substring(idx + 1).trim();
      output.push(`      ${JSON.stringify(key)}: ${JSON.stringify(value)},`);
    }
  });
  return output.join("\n");
});

fs.writeFileSync(pm2configPath, content);
