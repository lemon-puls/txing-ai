<template>
  <el-form :model="form" label-width="90px" class="project-edit-panel">
    <!-- ======== 基础信息（宽抽屉下两列排布，短字段并排、长文本独占整行） ======== -->
    <!-- basic info in two columns on the wide drawer: short fields side by side, long text full row -->
    <div class="group-title">
      <span class="gt-bar" />
      基础信息
    </div>
    <div class="basic-grid">
      <el-form-item label="项目名称">
        <el-input v-model="form.name" maxlength="30" placeholder="如 Txing-AI 智能对话平台" />
      </el-form-item>
      <el-form-item label="跳转链接">
        <el-input v-model="form.link" placeholder="https://...（选填，展示「访问项目」按钮）" />
      </el-form-item>
      <el-form-item label="项目描述" class="span-2">
        <el-input v-model="form.desc" type="textarea" :rows="2" maxlength="120" show-word-limit placeholder="卡片上的一句话简介" />
      </el-form-item>
      <el-form-item label="类别">
        <el-radio-group v-model="form.category">
          <el-radio-button value="company">公司项目</el-radio-button>
          <el-radio-button value="personal">个人项目</el-radio-button>
        </el-radio-group>
        <span class="form-tip">前台支持按公司/个人切换查看</span>
      </el-form-item>
      <el-form-item label="角标">
        <el-input v-model="form.badge" maxlength="10" placeholder="如 旗舰项目，展示在封面右上角" class="badge-input" />
      </el-form-item>
      <el-form-item label="回退渐变">
        <GradientPicker v-model="form.gradient" />
      </el-form-item>
      <el-form-item label="图标">
        <IconSelect v-model="form.iconKey" placeholder="无封面时的回退图标" />
      </el-form-item>
      <el-form-item label="标签" class="span-2">
        <TagsInput v-model="form.tags" add-text="添加标签" />
      </el-form-item>
      <el-form-item label="项目亮点" class="span-2">
        <el-input
          :model-value="(form.highlights || []).join('\n')"
          type="textarea"
          :rows="3"
          placeholder="每行一条亮点，前台以绿色对勾列表展示"
          @update:model-value="(v) => (form.highlights = v ? v.split('\n').map((s) => s.trim()).filter(Boolean) : [])"
        />
      </el-form-item>
    </div>

    <!-- ======== 媒体 ======== -->
    <div class="group-title">
      <span class="gt-bar" />
      媒体
    </div>
    <el-form-item label="封面轮播">
      <CoverMediaListEditor v-model="form.coverMedia" />
    </el-form-item>
    <el-form-item label="详情媒体">
      <MediaListEditor v-model="form.media" />
    </el-form-item>

    <!-- ======== 详情内容 ======== -->
    <div class="group-title">
      <span class="gt-bar" />
      详情内容
    </div>
    <el-form-item label="技术栈">
      <TechStackEditor v-model="form.techStack" />
    </el-form-item>
    <el-form-item label="核心功能">
      <FeatureEditor v-model="form.features" />
    </el-form-item>
    <el-form-item label="系统架构">
      <MarkdownInput
        v-model="form.architecture"
        :rows="8"
        placeholder="项目架构总览（选填，留空则前台不展示该区块）。支持 Markdown 与 ```mermaid 代码块（前台渲染为架构图），建议结构：```mermaid 架构图 + **分层说明** 列表"
      />
    </el-form-item>
  </el-form>
</template>

<script setup>
// 项目编辑表单体：三组字段（基础信息/媒体/详情内容），由 ProjectSection 的抽屉承载
// Project form body with three field groups; hosted by ProjectSection's drawer
import GradientPicker from './editors/GradientPicker.vue'
import IconSelect from './editors/IconSelect.vue'
import TagsInput from './editors/TagsInput.vue'
import CoverMediaListEditor from './editors/CoverMediaListEditor.vue'
import MediaListEditor from './editors/MediaListEditor.vue'
import TechStackEditor from './editors/TechStackEditor.vue'
import FeatureEditor from './editors/FeatureEditor.vue'
import MarkdownInput from './editors/MarkdownInput.vue'

// 表单对象由 Section 持有（open 时以详情接口重建），此处直接绑定字段
// the reactive form is owned by the section and rebuilt from the detail API on open
defineProps({
  form: { type: Object, required: true }
})
</script>

<style lang="scss" scoped>
.project-edit-panel {
  .group-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 22px 0 14px;
    font-size: 14px;
    font-weight: 600;
    color: var(--el-text-color-primary);

    &:first-child {
      margin-top: 0;
    }

    .gt-bar {
      width: 3px;
      height: 14px;
      border-radius: 2px;
      background: linear-gradient(180deg, var(--el-color-primary), #6366f1);
    }
  }

  // 基础信息两列栅格：短字段并排、长文本（描述/标签/亮点）span 整行；
  // 窄窗口（抽屉≈视口宽）回落单列，避免挤压
  // two-column grid for basic info; long-text fields span both columns;
  // collapses to one column on narrow viewports (drawer ≈ viewport width)
  .basic-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    column-gap: 24px;

    .span-2 {
      grid-column: span 2;
    }

    @media (max-width: 1080px) {
      grid-template-columns: minmax(0, 1fr);

      .span-2 {
        grid-column: auto;
      }
    }
  }

  .form-tip {
    margin-left: 12px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .badge-input {
    width: 200px;
  }
}
</style>
