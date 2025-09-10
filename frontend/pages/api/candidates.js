import { candidates } from "../../utils/mockApi";
export default function handler(req, res) {
  res.status(200).json({ candidates });
}
