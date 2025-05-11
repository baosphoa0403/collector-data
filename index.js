const express = require("express");
const bodyParser = require("body-parser");
const fs = require("fs");
const path = require("path");

const PORT = 3000;
const DATA_FILE = path.join(__dirname, "machines.json");

const app = express();
app.use(bodyParser.json());

// Đọc dữ liệu từ file JSON hoặc khởi tạo rỗng
function readData() {
  if (!fs.existsSync(DATA_FILE)) return {};
  try {
    const raw = fs.readFileSync(DATA_FILE, "utf-8");
    return JSON.parse(raw);
  } catch (e) {
    console.error("Error reading data:", e);
    return {};
  }
}

// Ghi dữ liệu vào file JSON
function writeData(data) {
  try {
    fs.writeFileSync(DATA_FILE, JSON.stringify(data, null, 2));
  } catch (e) {
    console.error("Error writing data:", e);
  }
}

// POST /api/report - nhận thông tin máy và upsert theo serial_number
app.post("/api/report", (req, res) => {
  const { serial_number, hostname, ip, ultra_id, time } = req.body;

  if (!serial_number) {
    return res.status(400).json({ error: "Missing serial_number" });
  }

  const db = readData();

  db[serial_number] = {
    serial_number, // 👈 lưu lại trong object
    hostname,
    ip,
    ultra_id,
    time,
  };

  writeData(db);
  console.log(`📝 Updated info for serial: ${serial_number}`);
  res.json({ message: "Upsert successful" });
});

// GET /api/machines - lấy toàn bộ danh sách
app.get("/api/machines", (req, res) => {
  res.json(readData());
});

// GET /api/machines/:serial - lấy 1 máy theo serial_number
app.get("/api/machines/:serial", (req, res) => {
  const db = readData();
  const serial = req.params.serial;
  if (db[serial]) {
    res.json(db[serial]);
  } else {
    res.status(404).json({ error: "Machine not found" });
  }
});

// Khởi động server
app.listen(PORT, () => {
  console.log(`🚀 API is running at http://localhost:${PORT}`);
});
