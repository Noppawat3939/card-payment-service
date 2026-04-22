# Webhook Simulator

Dev tool สำหรับจำลอง gateway webhook callback ไปยัง payment service

## Run

```bash
# จาก root ของ project
go run ./tools/webhook-simulator/...

# หรือผ่าน Makefile
make simulate
```

---

## Endpoints

| Method | Path              | Description                     |
| ------ | ----------------- | ------------------------------- |
| `GET`  | `/events`         | แสดง event ที่รองรับทั้งหมด     |
| `POST` | `/simulate`       | ยิง webhook 1 event             |
| `POST` | `/simulate/batch` | ยิง webhook หลาย event พร้อมกัน |
| `GET`  | `/health`         | health check                    |

---

## Available Events

| Event              | Description            |
| ------------------ | ---------------------- |
| `refund.completed` | Gateway คืนเงินสำเร็จ  |
| `refund.failed`    | Gateway คืนเงินล้มเหลว |
| `payment.success`  | Gateway charge สำเร็จ  |
| `payment.failed`   | Gateway charge ล้มเหลว |

---

## Usage

### ยิง single event

```bash
curl -X POST http://localhost:8081/simulate \
  -H 'Content-Type: application/json' \
  -d '{
    "event": "refund.completed",
    "gateway_ref": "gw_mock_001"
  }'
```

Response

```json
{
  "event": "refund.completed",
  "gateway_ref": "gw_mock_001",
  "status_code": 200,
  "success": true
}
```

---

### ยิง batch events

```bash
curl -X POST http://localhost:8081/simulate/batch \
  -H 'Content-Type: application/json' \
  -d '[
    { "event": "refund.completed", "gateway_ref": "gw_mock_001" },
    { "event": "payment.success",  "gateway_ref": "gw_mock_002" }
  ]'
```

---

### ดู available events

```bash
curl http://localhost:8081/events
```

---

## Makefile

เพิ่มใน `Makefile` ที่ root

```makefile
simulate:
	go run ./tools/webhook-simulator/...
```
