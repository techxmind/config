// 统一配置
// 从各种配置源融合成统一接口的配置
//  - 非静态配置的实时更新
//  - 自定义配置源插件
//  - 内容处理插件（加密配置，配置注释..)
// 配置使用应该遵循写少读多的原则，设计上为了保证并发读取性能，写入性能比较低
package config
