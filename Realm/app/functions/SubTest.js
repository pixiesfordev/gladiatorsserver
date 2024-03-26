exports = async function SubTest() {
  const ah = require("aura-gladiators");
  console.log("SubTest2")
  let myJson = {};
  try {
    myJson = await ah.GameJson.Get(ah.GameSetting.JsonName.Gladiator);
    console.log(JSON.stringify(myJson))
  } catch (error) {
    console.error("Error during getting Gladiator JSON:", error);
  }
  return JSON.stringify(ah.ReplyData.NewReplyData(myJson, null));

}
